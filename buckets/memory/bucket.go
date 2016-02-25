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
// Package memory presents an in-memory token bucket representation based on
// http://github.com/hotei/tokenbucket
package memory

import (
	"time"

	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/hotei/tokenbucket"
)

type bucketFactory struct {
}

func (bf *bucketFactory) Init(cfg *configs.ServiceConfig) {
	// A no-op
}

func (bf *bucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) buckets.Bucket {
	// fill rate is tokens-per-second.
	dur := time.Nanosecond * time.Duration(1e9 / cfg.FillRate)
	logging.Printf("Creating bucket for name %v with fill duration %v and capacity %v", buckets.FullyQualifiedName(namespace, bucketName), dur, cfg.Size)
	bucket := &tokenBucket{buckets.NewActivityChannel(), cfg, tokenbucket.New(dur, float64(cfg.Size))}
	return bucket
}

func NewBucketFactory() buckets.BucketFactory {
	return &bucketFactory{}
}

type tokenBucket struct {
	buckets.ActivityChannel
	cfg *configs.BucketConfig
	tb  *tokenbucket.TokenBucket
}

func (b *tokenBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	waitTime = b.tb.Take(numTokens)
	if waitTime > maxWaitTime && maxWaitTime > 0 {
		waitTime = -1
	}

	return
}

func (b *tokenBucket) Config() *configs.BucketConfig {
	return b.cfg
}
