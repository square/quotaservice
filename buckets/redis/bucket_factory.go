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
	flushedAtVersionKey       = "FlushedAtVersion"
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

	cfg               *pbconfig.ServiceConfig
	client            *redis.Client
	redisOpts         *redis.Options
	scriptSHA         string
	connectionRetries int
	flushdbCommand    string
}

// NewBucketFactory creates a new bucketFactory instance.
// flushdbCommand specifies the name of the Flushdb command. "flushdb"
// is used if it's empty.
func NewBucketFactory(redisOpts *redis.Options, connectionRetries int, flushdbCommand string) quotaservice.BucketFactory {
	if connectionRetries < 1 {
		connectionRetries = 1
	}
	if flushdbCommand == "" {
		flushdbCommand = "flushdb"
	}

	return &bucketFactory{
		redisOpts:         redisOpts,
		connectionRetries: connectionRetries,
		sharedAttributes:  make(map[string]*configAttributes),
		refcounts:         make(map[string]int),
		flushdbCommand:    flushdbCommand}
}

// Init initializes a bucketFactory for use, implementing Init() on the quotaservice.BucketFactory interface
func (bf *bucketFactory) Init(cfg *pbconfig.ServiceConfig) {
	logging.Printf("Initializing redis.bucketFactory for config version %v", cfg.Version)
	bf.Lock()
	defer bf.Unlock()

	bf.cfg = cfg

	if bf.client == nil {
		start := time.Now()
		bf.connectToRedisLocked()
		logging.Printf("Re-established Redis connections in %v", time.Since(start))
	}

	// Validate contents in Redis
	start := time.Now()
	// Check if Redis has been flushed when loading the current version of the configuration by a different node in
	// the cluster. If so, this node doesn't have to.
	val := bf.client.Get(flushedAtVersionKey)
	version, err := val.Int64()
	if err != nil {
		if err.Error() != "redis: nil" {
			// An actual problem; not just a missing flushedAtVersion.
			logging.Fatalf("Error response from Redis: %v", err)
		}

		// No flushedAtVersion has been established.
		logging.Print("flushedAtVersion not set")
		bf.flush(cfg.Version)

	} else {
		logging.Printf("flushedAtVersion is %v", version)
		if int32(version) < cfg.Version {
			bf.flush(cfg.Version)
		} else {
			logging.Printf("No need to flush since Redis has already been flushedAtVersion %v", version)
		}
	}
	logging.Printf("Verified Redis (including any flushes, if necessary) in %v", time.Since(start))
}

// flush flushes unused entries stored in Redis
func (bf *bucketFactory) flush(version int32) {
	start := time.Now()
	// Store flushedAtVersion to prevent multiple nodes flushing Redis unnecessarily, while the flush is in
	// progress. This may be racy, but it's a minor optimization. At worst case, we have > 1 flushes (from > 1
	// nodes in the cluster), which, while wasteful, isn't dangerous.
	if _, err := bf.client.Set(flushedAtVersionKey, version, 0).Result(); err != nil {
		logging.Printf("Failed to set flushedAtVersionKey: %v", err)
	}

	// We could consider a batched SCAN + DELETE if the FLUSHDB operation is slow. But for the most part, this is
	// "fast enough" - to the order of a few 100s of µs.
	if err := bf.client.Process(redis.NewStatusCmd(bf.flushdbCommand)); err != nil {
		logging.Printf("Failed to flushdb: %v", err)
	}

	// Re-set flushedAtVersion, since previous entry would have been removed with the flush.
	if _, err := bf.client.Set(flushedAtVersionKey, version, 0).Result(); err != nil {
		logging.Printf("Failed to set flushedAtVersionKey: %v", err)
	}
	logging.Printf("Flushed Redis in %v", time.Since(start))
}

func (bf *bucketFactory) connectToRedisLocked() {
	// Set up connection to Redis
	bf.client = redis.NewClient(bf.redisOpts)

	t, err := bf.client.Time().Result()
	if err != nil {
		logging.Printf("Cannot connect to Redis. TIME returned %v", err)
	} else {
		logging.Printf("Connection established. Time on Redis server: %v", t)
	}

	bf.scriptSHA = loadScript(bf.client)
}

func (bf *bucketFactory) reconnectToRedis(oldClient *redis.Client) {
	bf.Lock()
	defer bf.Unlock()

	// Always close connections on errors to prevent results leaking.
	if err := bf.client.Close(); unknownCloseError(err) {
		logging.Printf("Received error on Redis client close: %+v", err)
	}

	if oldClient == bf.client {
		bf.connectToRedisLocked()
	}
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

	keys := []string{toRedisKey(namespace, bucketName, tokensNextAvblNanosSuffix),
		toRedisKey(namespace, bucketName, accumulatedTokensSuffix)}

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

func toRedisKey(namespace, bucketName, suffix string) string {
	return namespace + ":" + bucketName + ":" + suffix
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
		logging.Fatalf("Unable to load LUA script into Redis; error=%v", s.Err())
		return
	}

	sha = s.Val()
	logging.Printf("Loaded LUA script into Redis; script SHA %v", sha)
	return
}
