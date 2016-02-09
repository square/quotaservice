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
	"gopkg.in/redis.v3"
	"sync"
)

type redisBucket struct {
	cfg    *configs.BucketConfig
	name   string
	client *redis.Client
}

type BucketFactory struct {
	*sync.RWMutex // Embed a lock.
	client      *redis.Client
	initialized bool
	buckets		[]*redisBucket
}

func (bf *BucketFactory) Init(cfg *configs.Configs) {
	bf.Lock()
	defer bf.Unlock()

	if !bf.initialized {
		bf.initialized = true

		// Set up connection to Redis
		// TODO(manik) read cfgs
		bf.client = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0, // use default DB
		})

		// Set up buckets slice
		bf.buckets = make([]*redisBucket)
	}

	go bf.runReplenisher()
}

func (bf *BucketFactory) runReplenisher() {
	for _ := range time.Tick(time.Millisecond * 100) {
		bf.refillBuckets()
	}
}

func (bf *BucketFactory) refillBuckets() {
	bf.RLock()
	defer bf.Unlock()

	for rb := range bf.buckets {
		// TODO(manik)
	}
}

func (bf *BucketFactory) NewBucket(name string, cfg *configs.BucketConfig) buckets.Bucket {
	return &redisBucket{cfg: cfg, name: name, client: bf.client}
}

func (b *redisBucket) TakeBlocking(numTokens int, timeout time.Duration) (success bool) {
	m := b.client.Multi()
	defer m.Exec(nil)
	deadline := time.Now() + timeout
	for ; deadline > time.Now(); {
		s := m.Get(b.name)
		tokens, err := s.Int64()
		if err != nil {
			return false
		}
		if tokens > numTokens {
			m.DecrBy(b.name, numTokens)
			return true
		} else {
			// TODO(manik) use some inter-thread signalling
			time.Sleep(10 * time.Millisecond)
		}
	}
	return false
}

func (b *redisBucket) Take(numTokens int) (success bool) {
	m := b.client.Multi()
	defer m.Exec(nil)
	s := m.Get(b.name)
	tokens, err := s.Int64()
	if err != nil {
		return false
	}
	if tokens > numTokens {
		m.DecrBy(b.name, numTokens)
		return true
	} else {
		return false
	}
}

func (b *redisBucket) GetConfig() *configs.BucketConfig {
	return b.cfg
}
