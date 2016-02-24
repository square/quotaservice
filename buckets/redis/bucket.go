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

const (
	TOKENS_NEXT_AVBL_NANOS_SUFFIX = "TNA"
	ACCUMULATED_TOKENS_SUFFIX = "AT"
)

type redisBucket struct {
	cfg                   *configs.BucketConfig
	namespace             string
	name                  string
	factory               *BucketFactory
	nanosBetweenTokens    string
	maxTokensToAccumulate string
	redisKeys             []string // {tokensNextAvailableRedisKey, accumulatedTokensRedisKey}
	buckets.ActivityChannel
}

type BucketFactory struct {
	m           *sync.RWMutex // Embed a lock.
	client      *redis.Client
	initialized bool
	redisOpts   *redis.Options
	scriptSHA   string
}

func NewBucketFactory(redisOpts *redis.Options) *BucketFactory {
	return &BucketFactory{initialized: false, m: &sync.RWMutex{}, redisOpts: redisOpts}
}

func (bf *BucketFactory) Init(cfg *configs.ServiceConfig) {
	if !bf.initialized {
		bf.m.Lock()
		defer bf.m.Unlock()

		if !bf.initialized {
			bf.initialized = true
			logging.Print("Establishing connection to Redis")
			// Set up connection to Redis
			bf.client = redis.NewClient(bf.redisOpts)
			logging.Printf("Connection established. Time on server: %v", time.Unix(toInt64(bf.client.Time().Val()[0], 0) / 1000, 0))
			bf.scriptSHA = loadScript(bf.client)
		}
	}
}

func (bf *BucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) buckets.Bucket {
	return &redisBucket{
		cfg,
		namespace,
		bucketName,
		bf,
		strconv.FormatInt(int64(1e9) / int64(cfg.FillRate), 10),
		strconv.FormatInt(int64(cfg.Size), 10),
		[]string{toRedisKey(namespace, bucketName, TOKENS_NEXT_AVBL_NANOS_SUFFIX),
			toRedisKey(namespace, bucketName, ACCUMULATED_TOKENS_SUFFIX)},
		buckets.NewActivityChannel()}
}

func toRedisKey(namespace, bucketName, suffix string) string {
	return fmt.Sprintf("%v:%v:%v", namespace, bucketName, suffix)
}

func (b *redisBucket) Take(requested int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	currentTimeNanos := fmt.Sprintf("%v", time.Now().UnixNano())
	args := []string{currentTimeNanos, b.nanosBetweenTokens, b.maxTokensToAccumulate,
		strconv.FormatInt(requested, 10), strconv.FormatInt(maxWaitTime.Nanoseconds(), 10)}

	res := b.factory.client.EvalSha(b.factory.scriptSHA, b.redisKeys, args)
	waitTimeNanos := res.Val().(int64)
	waitTime = time.Nanosecond * time.Duration(waitTimeNanos)
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

func loadScript(c *redis.Client) (sha string) {
	lua := `
	local zero = tonumber("0")
	local tokensNextAvailableNanos = tonumber(redis.call("GET", KEYS[1]))
	if not tokensNextAvailableNanos then
		tokensNextAvailableNanos = zero
	end

	local accumulatedTokens = redis.call("GET", KEYS[2])
	if not accumulatedTokens then
		accumulatedTokens = zero
	end

	local currentTimeNanos = tonumber(ARGV[1])
	local nanosBetweenTokens = tonumber(ARGV[2])
	local maxTokensToAccumulate = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])
	local maxWaitTime = tonumber(ARGV[5])
	local freshTokens = zero

	if currentTimeNanos > tokensNextAvailableNanos then
		freshTokens = math.floor((currentTimeNanos - tokensNextAvailableNanos) / nanosBetweenTokens)
		accumulatedTokens = math.min(maxTokensToAccumulate, accumulatedTokens + freshTokens)
		tokensNextAvailableNanos = currentTimeNanos
	end

	local waitTime = tokensNextAvailableNanos - currentTimeNanos
	local accumulatedTokensUsed = math.min(accumulatedTokens, requested)
	local tokensToWaitFor = requested - accumulatedTokensUsed
	local futureWaitNanos = tokensToWaitFor * nanosBetweenTokens

	if waitTime > 0 and waitTime > maxWaitTime and maxWaitTime > 0 then
    	waitTime = -1
	else
		tokensNextAvailableNanos = tokensNextAvailableNanos + futureWaitNanos
		accumulatedTokens = accumulatedTokens - accumulatedTokensUsed

		redis.call("SET", KEYS[1], tokensNextAvailableNanos)
		redis.call("SET", KEYS[2], math.floor(accumulatedTokens))
	end

	return waitTime
	`
	s := c.ScriptLoad(lua)
	sha = s.Val()
	logging.Printf("Loaded LUA script into Redis; script SHA %v", sha)
	return
}
