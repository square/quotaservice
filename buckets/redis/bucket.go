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
	cfg                         *configs.BucketConfig
	namespace                   string
	name                        string
	factory                     *BucketFactory
	nanosBetweenTokens          int64
	maxTokensToAccumulate       int64
	tokensNextAvblNanosRedisKey string
	accumulatedTokensRedisKey   string
}

type BucketFactory struct {
	m           *sync.RWMutex // Embed a lock.
	client      *redis.Client
	initialized bool
}

func NewBucketFactory() *BucketFactory {
	return &BucketFactory{initialized: false, m: &sync.RWMutex{}}
}

func (bf *BucketFactory) Init(cfg *configs.ServiceConfig) {
	if !bf.initialized {
		bf.m.Lock()
		defer bf.m.Unlock()

		if !bf.initialized {
			bf.initialized = true
			logging.Print("Establishing connection to Redis")
			// Set up connection to Redis
			// TODO(manik) read cfgs from config
			bf.client = redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "", // no password set
				DB:       0, // use default DB
			})

			logging.Printf("Connection established. Time on server: %v", time.Unix(toInt64(bf.client.Time().Val()[0], 0) / 1000, 0))
		}
	}
}

func (bf *BucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) buckets.Bucket {
	return &redisBucket{
		cfg: cfg,
		namespace: namespace,
		name: bucketName,
		factory: bf,
		nanosBetweenTokens: int64(1e9 / cfg.FillRate),
		maxTokensToAccumulate: int64(cfg.Size),
		tokensNextAvblNanosRedisKey: toRedisKey(namespace, bucketName, TOKENS_NEXT_AVBL_NANOS_SUFFIX),
		accumulatedTokensRedisKey: toRedisKey(namespace, bucketName, ACCUMULATED_TOKENS_SUFFIX)}
}

func toRedisKey(namespace, bucketName, suffix string) string {
	return fmt.Sprintf("%v:%v:%v", namespace, bucketName, suffix)
}

func (b *redisBucket) Take(requested int, maxWaitTime time.Duration) (waitTime time.Duration) {
	tokensRequested := int64(requested)
	// Start a Redis "transaction"
	client := b.factory.client
	m := client.Multi()
	defer m.Exec(func() error {
		return nil
	})

	currentTimeNanos := time.Now().UnixNano()
	cachedVals := m.MGet(b.tokensNextAvblNanosRedisKey, b.accumulatedTokensRedisKey)

	tokensNextAvailableNanos := toInt64(cachedVals.Val()[0], 0)
	accumulatedTokens := toInt64(cachedVals.Val()[1], 0)

	if currentTimeNanos > tokensNextAvailableNanos {
		// Accumulate fresh tokens.
		freshTokens := (currentTimeNanos - tokensNextAvailableNanos) / b.nanosBetweenTokens

		// Never exceed maxTokensToAccumulate
		accumulatedTokens = min(b.maxTokensToAccumulate, accumulatedTokens + freshTokens)
		tokensNextAvailableNanos = currentTimeNanos
	}

	waitTime = time.Duration(tokensNextAvailableNanos - currentTimeNanos) * time.Nanosecond
	accumulatedTokensUsed := min(accumulatedTokens, tokensRequested)
	tokensToWaitFor := int64(tokensRequested) - accumulatedTokensUsed
	futureWaitNanos := tokensToWaitFor * b.nanosBetweenTokens

	// Is waitTime too long?
	if waitTime > 0 && waitTime > maxWaitTime && maxWaitTime > 0 {
		// Don't "claim" any tokens.
		waitTime = time.Duration(-1)
	} else {
		tokensNextAvailableNanos = tokensNextAvailableNanos + futureWaitNanos
		accumulatedTokens = accumulatedTokens - accumulatedTokensUsed

		// TODO(manik) see if we can re-implement using INCR?

		// "Claim" tokens by writing variables back to Redis and releasing lock.
		m.Set(b.tokensNextAvblNanosRedisKey, tokensNextAvailableNanos, 0)
		m.Set(b.accumulatedTokensRedisKey, accumulatedTokens, 0)
	}

	return
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
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

func (b *redisBucket) GetConfig() *configs.BucketConfig {
	return b.cfg
}
