/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */
// Package redis implements token buckets backed by Redis, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package redis

import (
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/buckets"
	"time"
	"sync"
	"gopkg.in/redis.v3"
	"fmt"
	"github.com/maniksurtani/quotaservice/logging"
	"strconv"
)

// Suffixes for Redis keys
const (
	TOKENS_NEXT_AVBL_NANOS_SUFFIX = "TNA"
	ACCUMULATED_TOKENS_SUFFIX = "AT"
)

// redisBucket is threadsafe since it delegates concurrency to the Redis instance.
type redisBucket struct {
	dynamic               bool
	cfg                   *configs.BucketConfig
	factory               *bucketFactory
	nanosBetweenTokens    string
	maxTokensToAccumulate string
	maxIdleTimeMillis     string
	maxDebtNanos          string
	redisKeys             []string // {tokensNextAvailableRedisKey, accumulatedTokensRedisKey}
	buckets.ActivityChannel
}

type bucketFactory struct {
	cfg               *configs.ServiceConfig
	m                 *sync.RWMutex
	client            *redis.Client
	initialized       bool
	redisOpts         *redis.Options
	scriptSHA         string
	connectionRetries int
}

func NewBucketFactory(redisOpts *redis.Options, connectionRetries int) buckets.BucketFactory {
	if connectionRetries < 1 {
		connectionRetries = 1
	}

	return &bucketFactory{
		initialized: false,
		m: &sync.RWMutex{},
		redisOpts: redisOpts,
		connectionRetries: connectionRetries}
}

func (bf *bucketFactory) Init(cfg *configs.ServiceConfig) {
	if !bf.initialized {
		bf.m.Lock()
		defer bf.m.Unlock()

		if !bf.initialized {
			bf.initialized = true
			bf.cfg = cfg
			bf.connectToRedis()
		}
	}
}

func (bf *bucketFactory) connectToRedis() {
	// Set up connection to Redis
	bf.client = redis.NewClient(bf.redisOpts)
	logging.Printf("Connection established. Time on Redis server: %v", time.Unix(toInt64(bf.client.Time().Val()[0], 0), 0))
	bf.scriptSHA = loadScript(bf.client)
}

func (bf *bucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig, dyn bool) buckets.Bucket {
	idle := "0"
	if cfg.MaxIdleMillis > 0 {
		idle = strconv.FormatInt(int64(cfg.MaxIdleMillis), 10)
	}

	rb := &redisBucket{
		dyn,
		cfg,
		bf,
		strconv.FormatInt(1e9 / cfg.FillRate, 10),
		strconv.FormatInt(cfg.Size, 10),
		idle,
		strconv.FormatInt(cfg.MaxDebtMillis * 1e6, 10), // Convert millis to nanos
		[]string{toRedisKey(namespace, bucketName, TOKENS_NEXT_AVBL_NANOS_SUFFIX),
			toRedisKey(namespace, bucketName, ACCUMULATED_TOKENS_SUFFIX)},
		buckets.NewActivityChannel()}

	return rb
}

func toRedisKey(namespace, bucketName, suffix string) string {
	return fmt.Sprintf("%v:%v:%v", namespace, bucketName, suffix)
}

func (b *redisBucket) Take(requested int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	currentTimeNanos := strconv.FormatInt(time.Now().UnixNano(), 10)
	args := []string{currentTimeNanos, b.nanosBetweenTokens, b.maxTokensToAccumulate,
		strconv.FormatInt(requested, 10), strconv.FormatInt(maxWaitTime.Nanoseconds(), 10),
		b.maxIdleTimeMillis, b.maxDebtNanos}

	keepTrying := true
	for attempt := 0; keepTrying && attempt < b.factory.connectionRetries; attempt++ {
		res := b.factory.client.EvalSha(b.factory.scriptSHA, b.redisKeys, args)
		switch waitTimeNanos := res.Val().(type) {
		case int64:
			waitTime = time.Nanosecond * time.Duration(waitTimeNanos)
			keepTrying = false
		default:
			if res.Err() != nil && res.Err().Error() == "redis: client is closed" {
				b.factory.connectToRedis()
			} else {
				keepTrying = false
				panic(fmt.Sprintf("Unknown response '%v' of type %T. Full result %+v",
					waitTimeNanos, waitTimeNanos, res))
			}
		}
	}

	if keepTrying {
		panic(fmt.Sprintf("Couldn't reconnect to Redis, even after %v attempts",
			b.factory.connectionRetries))
	}

	return
}

func toInt64(s interface{}, defaultValue int64) (v int64) {
	if s != nil {
		var err error
		v, err = strconv.ParseInt(s.(string), 10, 64)
		if err != nil {
			logging.Printf("Cannot convert '%v' to int64", s)
		}
	}
	return
}

func (b *redisBucket) Config() *configs.BucketConfig {
	return b.cfg
}

func (b *redisBucket) Dynamic() bool {
	return b.dynamic
}

func (b *redisBucket) Destroy() {
	// No-op
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

	if (tokensNextAvailableNanos - currentTimeNanos > maxDebtNanos) or (waitTime > 0 and waitTime > maxWaitTime and maxWaitTime > 0) then
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
