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

type BucketFactory struct {
}

func (bf BucketFactory) Init(cfg *configs.ServiceConfig) {
	// A no-op
}

func (bf BucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) buckets.Bucket {
	dur := time.Nanosecond * time.Duration(1000000000 / cfg.FillRate)
	logging.Printf("Creating bucket for name %v:%v with fill duration %v and capacity %v", namespace, bucketName, dur, cfg.Size)
	tb := tokenbucket.New(dur, float64(cfg.Size))
	return &tokenBucket{cfg: cfg, tb: tb}
}

type tokenBucket struct {
	tb  *tokenbucket.TokenBucket // Embed actual token bucket
	cfg *configs.BucketConfig
}

func (b *tokenBucket) Take(numTokens int, maxWaitTime time.Duration) (waitTime time.Duration) {
	// TODO(manik) respect maxWaitTime
	return b.tb.Take(int64(numTokens))
}

func (b *tokenBucket) GetConfig() *configs.BucketConfig {
	return b.cfg
}
