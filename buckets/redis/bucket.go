// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package redis implements token buckets backed by Redis, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package redis

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v5"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/logging"

	"sync"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

// Suffixes for Redis keys
const (
	tokensNextAvblNanosSuffix = "TNA"
	accumulatedTokensSuffix   = "AT"
)

// redisBucket is threadsafe since it delegates concurrency to the Redis instance.
type redisBucket struct {
	dynamic                    bool
	cfg                        *pbconfig.BucketConfig
	factory                    *bucketFactory
	nanosBetweenTokens         string
	maxTokensToAccumulate      string
	maxIdleTimeMillis          string
	maxDebtNanos               string
	redisKeys                  []string // {tokensNextAvailableRedisKey, accumulatedTokensRedisKey}
	quotaservice.DefaultBucket          // Extension for default methods on interface
}

type bucketFactory struct {
	cfg               *pbconfig.ServiceConfig
	client            *redis.Client
	redisOpts         *redis.Options
	scriptSHA         string
	connectionRetries int
	mu                sync.Mutex
}

func NewBucketFactory(redisOpts *redis.Options, connectionRetries int) quotaservice.BucketFactory {
	if connectionRetries < 1 {
		connectionRetries = 1
	}

	return &bucketFactory{
		redisOpts:         redisOpts,
		connectionRetries: connectionRetries}
}

func (bf *bucketFactory) Init(cfg *pbconfig.ServiceConfig) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	bf.cfg = cfg

	if bf.client == nil {
		bf.connectToRedisLocked()
	}

	// Clean out all buckets, start afresh.
	bf.client.FlushDb()
}

func (bf *bucketFactory) connectToRedisLocked() {
	// Set up connection to Redis
	bf.client = redis.NewClient(bf.redisOpts)

	time, err := bf.client.Time().Result()
	if err != nil {
		logging.Printf("Cannot connect to Redis. TIME returned %v", err)
	} else {
		logging.Printf("Connection established. Time on Redis server: %v", time)
	}

	bf.scriptSHA = loadScript(bf.client)
}

func (bf *bucketFactory) reconnectToRedis(oldClient *redis.Client) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if oldClient == bf.client {
		bf.connectToRedisLocked()
	}
}

func (bf *bucketFactory) Client() interface{} {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	return bf.client
}

func (bf *bucketFactory) NewBucket(namespace, bucketName string, cfg *pbconfig.BucketConfig, dyn bool) quotaservice.Bucket {
	idle := "0"
	if cfg.MaxIdleMillis > 0 {
		idle = strconv.FormatInt(int64(cfg.MaxIdleMillis), 10)
	}

	rb := &redisBucket{
		dyn,
		cfg,
		bf,
		strconv.FormatInt(1e9/cfg.FillRate, 10),
		strconv.FormatInt(cfg.Size, 10),
		idle,
		strconv.FormatInt(cfg.MaxDebtMillis*1e6, 10), // Convert millis to nanos
		[]string{toRedisKey(namespace, bucketName, tokensNextAvblNanosSuffix),
			toRedisKey(namespace, bucketName, accumulatedTokensSuffix)},
		*new(quotaservice.DefaultBucket)}

	return rb
}

func toRedisKey(namespace, bucketName, suffix string) string {
	return namespace + ":" + bucketName + ":" + suffix
}

func (b *redisBucket) Take(requested int64, maxWaitTime time.Duration) (time.Duration, bool) {
	currentTimeNanos := strconv.FormatInt(time.Now().UnixNano(), 10)
	args := []interface{}{currentTimeNanos, b.nanosBetweenTokens, b.maxTokensToAccumulate,
		strconv.FormatInt(requested, 10), strconv.FormatInt(maxWaitTime.Nanoseconds(), 10),
		b.maxIdleTimeMillis, b.maxDebtNanos}

	keepTrying := true
	var waitTime time.Duration
	for attempt := 0; keepTrying && attempt < b.factory.connectionRetries; attempt++ {
		client := b.factory.Client().(*redis.Client)
		res := client.EvalSha(b.factory.scriptSHA, b.redisKeys, args...)
		switch waitTimeNanos := res.Val().(type) {
		case int64:
			waitTime = time.Nanosecond * time.Duration(waitTimeNanos)
			keepTrying = false
		default:
			// Always close connections on errors to prevent results leaking.
			if err := b.factory.client.Close(); err != nil {
				logging.Printf("Received error on redis client close: %+v", err)
			}

			if res.Err() != nil && res.Err().Error() == "redis: client is closed" {
				b.factory.reconnectToRedis(client)
			} else {
				logging.Printf("Unknown response '%v' of type %T. Full result %+v",
					waitTimeNanos, waitTimeNanos, res)
				b.factory.reconnectToRedis(client)
			}
		}
	}

	if keepTrying {
		panic(fmt.Sprintf("Couldn't reconnect to Redis, even after %v attempts",
			b.factory.connectionRetries))
	}

	if waitTime < 0 {
		// Timed out
		return 0, false
	}

	return waitTime, true
}

func (b *redisBucket) Config() *pbconfig.BucketConfig {
	return b.cfg
}

func (b *redisBucket) Dynamic() bool {
	return b.dynamic
}

func checkScriptExists(c *redis.Client, sha string) bool {
	r := c.ScriptExists(sha)
	return r.Val()[0]
}

// loadScript loads the LUA script into Redis. The LUA script contains the token bucket algorithm
// which is executed atomically in Redis. Once the script is loaded, it is invoked using its SHA.
func loadScript(c *redis.Client) (sha string) {
	lua := `
	local tokensNextAvailableNanos = tonumber(redis.call("GET", KEYS[1]))
	if not tokensNextAvailableNanos then
		tokensNextAvailableNanos = 0
	end

	local maxTokensToAccumulate = tonumber(ARGV[3])

	local accumulatedTokens = redis.call("GET", KEYS[2])
	if not accumulatedTokens then
		accumulatedTokens = maxTokensToAccumulate
	end

	local currentTimeNanos = tonumber(ARGV[1])
	local nanosBetweenTokens = tonumber(ARGV[2])
	local requested = tonumber(ARGV[4])
	local maxWaitTime = tonumber(ARGV[5])
	local lifespan = tonumber(ARGV[6])
	local maxDebtNanos = tonumber(ARGV[7])
	local freshTokens = 0

	if currentTimeNanos > tokensNextAvailableNanos then
		freshTokens = math.floor((currentTimeNanos - tokensNextAvailableNanos) / nanosBetweenTokens)
		accumulatedTokens = math.min(maxTokensToAccumulate, accumulatedTokens + freshTokens)
		tokensNextAvailableNanos = currentTimeNanos
	end

	local waitTime = tokensNextAvailableNanos - currentTimeNanos
	local accumulatedTokensUsed = math.min(accumulatedTokens, requested)
	local tokensToWaitFor = requested - accumulatedTokensUsed
	local futureWaitNanos = tokensToWaitFor * nanosBetweenTokens

	tokensNextAvailableNanos = tokensNextAvailableNanos + futureWaitNanos
	accumulatedTokens = accumulatedTokens - accumulatedTokensUsed

	if (tokensNextAvailableNanos - currentTimeNanos > maxDebtNanos) or (waitTime > 0 and waitTime > maxWaitTime) then
    	waitTime = -1
	else
		if lifespan > 0 then
			redis.call("SET", KEYS[1], tokensNextAvailableNanos, "PX", lifespan)
			redis.call("SET", KEYS[2], math.floor(accumulatedTokens), "PX", lifespan)
		else
			redis.call("SET", KEYS[1], tokensNextAvailableNanos)
			redis.call("SET", KEYS[2], math.floor(accumulatedTokens))
		end
	end

	return waitTime
	`
	s := c.ScriptLoad(lua)
	sha = s.Val()
	logging.Printf("Loaded LUA script into Redis; script SHA %v", sha)
	return
}
