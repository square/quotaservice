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
)

type redisBucket struct {
	cfg     *configs.BucketConfig
	name    string
	factory *BucketFactory
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

			// Set up connection to Redis
			// TODO(manik) read cfgs
			bf.client = redis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "", // no password set
				DB:       0, // use default DB
			})
		}
	}
}

func (bf *BucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) buckets.Bucket {
	return &redisBucket{cfg: cfg, name: bucketName, factory: bf}
}

func (b *redisBucket) Take(numTokens int, maxWaitTime time.Duration) (waitTime time.Duration) {

	client := b.factory.client
	m := client.Multi()
	defer m.Exec(nil)
	s := m.Get(b.name)
	tokens, err := s.Int64()
	if err != nil {
		return 0
	}
	if tokens > int64(numTokens) {
		m.DecrBy(b.name, int64(numTokens))
		return 0
	} else {
		return 0
	}
	return 0
}

func (b *redisBucket) GetConfig() *configs.BucketConfig {
	return b.cfg
}
