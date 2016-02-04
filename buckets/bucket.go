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

package buckets
import (
	"time"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice/configs"
)

type TokenBucketsContainer struct {
	buckets map[string]Bucket
}

type Bucket interface {
	TakeBlocking(numTokens int, timeout time.Duration) (success bool)
	Take(numTokens int) (success bool)
	GetConfig() *configs.BucketConfig
}

type BucketFactory interface {
	Init(cfg *configs.Configs)
	NewBucket(name string, cfg *configs.BucketConfig) Bucket
}

var tokenBuckets *TokenBucketsContainer

func InitBuckets(cfg *configs.Configs, bf BucketFactory) *TokenBucketsContainer {
	tokenBuckets = &TokenBucketsContainer{buckets:make(map[string]Bucket)}
	logging.Print("Initializing buckets")
	bf.Init(cfg)
	for n, b := range cfg.Buckets {
		tokenBuckets.buckets[n] = bf.NewBucket(n, b)
	}
	logging.Print("Finished initializing buckets")
	return tokenBuckets
}

func (tb *TokenBucketsContainer) FindBucket(name string) Bucket {
	b := tb.buckets[name]
	// TODO(manik) perform an actual search
	return b
}
