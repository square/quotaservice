// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package redis implements token buckets backed by Redis, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package redis

import (
	"strconv"
	"time"

	"gopkg.in/redis.v5"

	"github.com/square/quotaservice"
	"github.com/square/quotaservice/logging"

	"sync"

	pbconfig "github.com/square/quotaservice/protos/config"
)

// Suffixes for Redis keys
const (
	tokensNextAvblNanosSuffix = "TNA"
	accumulatedTokensSuffix   = "AT"
)

// defaultBucket is a "const"
var defaultBucket = &quotaservice.DefaultBucket{}

// bucketFactory holds an instance of the Redis client, and constructs staticBucket and dynamicBucket instances for use
// with Redis. Contains an embedded mutex which should be used when reading or updating the reference to the Redis
// client. Also holds references to configAttributes for each namespace and refcounts of usage of commonAttributes,
// both also guarded by this mutex.
type bucketFactory struct {
	// Embedded mutex
	sync.Mutex

	// Refcounts of configAttributes instances used by dynamic buckets for each namespace, protected by the
	// embedded mutex.
	refcounts map[string]int

	// sharedAttributes are instances of configAttributes used by dynamic buckets for each namespace, protected by
	// the embedded mutex.
	sharedAttributes map[string]*configAttributes

	cfg                       *pbconfig.ServiceConfig
	client                    *redis.Client
	redisOpts                 *redis.Options
	scriptSHA                 string
	connectionRetries         int
	connectionNeedsResolution bool
	numTimesConnResolved      int // For testing and debugging purposes

	// keyMaxIdleTime will be set as the Redis key TTL unless it is overridden by the per bucket
	// config MaxIdleMillis
	keyMaxIdleTime time.Duration
}

// NewBucketFactory creates a new bucketFactory instance.
func NewBucketFactory(redisOpts *redis.Options, connectionRetries int, keyMaxIdleTime time.Duration) quotaservice.BucketFactory {
	if connectionRetries < 1 {
		connectionRetries = 1
	}

	if keyMaxIdleTime == 0 {
		keyMaxIdleTime = 24 * time.Hour
	}

	return &bucketFactory{
		redisOpts:                 redisOpts,
		connectionRetries:         connectionRetries,
		sharedAttributes:          make(map[string]*configAttributes),
		refcounts:                 make(map[string]int),
		connectionNeedsResolution: false,
		numTimesConnResolved:      0,
		keyMaxIdleTime:            keyMaxIdleTime,
	}
}

// Init initializes a bucketFactory for use, implementing Init() on the quotaservice.BucketFactory interface
func (bf *bucketFactory) Init(cfg *pbconfig.ServiceConfig) {
	start := time.Now()
	logging.Infof("Initializing redis.bucketFactory for config version %v", cfg.Version)
	bf.Lock()
	defer bf.Unlock()

	bf.cfg = cfg

	if bf.client == nil {
		connStart := time.Now()
		bf.connectToRedisLocked()
		logging.Debug("Re-established Redis connections in %v", time.Since(connStart))
	}

	logging.Infof("Initialized redis.BucketFactory in %v", time.Since(start))
}

func (bf *bucketFactory) connectToRedisLocked() {
	// Set up connection to Redis
	bf.client = redis.NewClient(bf.redisOpts)

	t, err := bf.client.Time().Result()
	if err != nil {
		logging.Errorf("Cannot connect to Redis. TIME returned %v", err)
	} else {
		logging.Debugf("Connection established. Time on Redis server: %v", t)
	}

	bf.scriptSHA = loadScript(bf.client)
}

func (bf *bucketFactory) reconnectToRedis(oldClient *redis.Client) {
	bf.Lock()
	defer bf.Unlock()

	// Always close connections on errors to prevent results leaking.
	if err := bf.client.Close(); unknownCloseError(err) {
		logging.Warnf("Received error on Redis client close: %+v", err)
	}

	if oldClient == bf.client {
		bf.connectToRedisLocked()
	}
}

func (bf *bucketFactory) handleConnectionFailure(oldClient *redis.Client) {
	bf.Lock()
	defer bf.Unlock()

	if oldClient == bf.client && !bf.connectionNeedsResolution {
		logging.Debug("Attempting to establish new connection to redis")
		bf.connectionNeedsResolution = true
		go bf.establishNewConnectionToRedis(oldClient)
	}
}

func (bf *bucketFactory) establishNewConnectionToRedis(oldClient *redis.Client) {
	client := oldClient
	numsTried := 0
	exponentialDelay := 1 * time.Second
	exponentialDelayCeiling := 3 * time.Minute

	for disconnected := true; disconnected; {
		attemptsRemaining := bf.connectionRetries

		for attemptsRemaining > 0 {
			attemptsRemaining--
			numsTried++
			bf.reconnectToRedis(client)

			client = bf.Client().(*redis.Client)
			_, err := client.Ping().Result()
			if err == nil {
				disconnected = false
				break
			}
			logging.Trace("Unable to reconnect to redis. Will retry again in 1s.")
			time.Sleep(1 * time.Second)
		}

		if disconnected {
			logging.Tracef("Attempted to reconnect %v times. Will sleep for %v seconds", bf.connectionRetries, exponentialDelay)
			time.Sleep(exponentialDelay)
			exponentialDelay *= 2

			if exponentialDelay > exponentialDelayCeiling {
				logging.Tracef("Resetting exponential delay for sleep because it exceeds the ceiling value of %v seconds", exponentialDelayCeiling)
				exponentialDelay = 1 * time.Second
			}
		}

	}

	logging.Debugf("Established connection after attempting %v times", numsTried)
	bf.Lock()
	defer bf.Unlock()
	bf.connectionNeedsResolution = false
	bf.numTimesConnResolved++
	exponentialDelay = 1 * time.Second
	logging.Debugf("Handler has resolved %v connection(s) so far", bf.numTimesConnResolved)
}

func (bf *bucketFactory) getNumTimesConnResolved() int {
	bf.Lock()
	defer bf.Unlock()

	return bf.numTimesConnResolved
}

// Client returns a reference to the underlying client instance, implementing Client() on the quotaservice.BucketFactory
// interface
func (bf *bucketFactory) Client() interface{} {
	bf.Lock()
	defer bf.Unlock()

	return bf.client
}

// NewBucket creates and returns a new instance of quotaservice.Bucket, implementing NewBucket() on the
// quotaservice.BucketFactory interface
func (bf *bucketFactory) NewBucket(namespace, bucketName string, cfg *pbconfig.BucketConfig, dyn bool) quotaservice.Bucket {
	idle := "0"
	if cfg.MaxIdleMillis > 0 {
		idle = strconv.FormatInt(int64(cfg.MaxIdleMillis), 10)
	}

	keys := []string{toRedisKey(namespace, bucketName, tokensNextAvblNanosSuffix, bf.cfg.Version),
		toRedisKey(namespace, bucketName, accumulatedTokensSuffix, bf.cfg.Version)}

	if dyn {
		var exists bool
		bf.Lock()
		defer bf.Unlock()

		var attribs *configAttributes

		if attribs, exists = bf.sharedAttributes[namespace]; !exists {
			attribs = newConfigAttributes(cfg, idle, dyn)
			bf.sharedAttributes[namespace] = attribs
			bf.refcounts[namespace] = 0
		}
		bf.refcounts[namespace]++

		// Create a dynamicBucket with a reference to the appropriate shared configAttributes instance
		return &dynamicBucket{
			&abstractBucket{
				attribs,
				cfg,
				bf,
				keys}}
	} else {
		// Create a staticBucket with its own non-shared configAttributes
		return &staticBucket{
			&abstractBucket{
				newConfigAttributes(cfg, idle, dyn),
				cfg,
				bf,
				keys}}
	}
}

func newConfigAttributes(cfg *pbconfig.BucketConfig, idle string, dyn bool) *configAttributes {
	return &configAttributes{
		strconv.FormatInt(1e9/cfg.FillRate, 10),
		strconv.FormatInt(cfg.Size, 10),
		idle,
		// Convert millis to nanos
		strconv.FormatInt(cfg.MaxDebtMillis*1e6, 10),
		defaultBucket}
}

func toRedisKey(namespace, bucketName, suffix string, version int32) string {
	return namespace + ":" + bucketName + ":" + suffix + ":" + strconv.Itoa(int(version))
}

func unknownCloseError(err error) bool {
	return err != nil && err.Error() != "redis: client is closed"
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
	if s.Err() != nil {
		logging.Errorf("Unable to load LUA script into Redis; error=%v", s.Err())
		return
	}

	sha = s.Val()
	logging.Debugf("Loaded LUA script into Redis; script SHA %v", sha)
	return
}
