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
package test

import (
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/buckets/redis"
	"testing"
	"os"
	"github.com/maniksurtani/quotaservice/configs"
	"fmt"
	"time"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	r "gopkg.in/redis.v3"
)

var factories = map[string]buckets.BucketFactory{
	"memory": memory.NewBucketFactory(),
	"redis": redis.NewBucketFactory(&r.Options{Addr: "localhost:6379"}, 2)}

var testBuckets map[string]buckets.Bucket = make(map[string]buckets.Bucket)

var cfg = configs.NewDefaultServiceConfig()

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	for impl, factory := range factories {
		factory.Init(cfg)
		fullyQualifiedName := buckets.FullyQualifiedName(impl, impl)
		testBuckets[fullyQualifiedName] = factory.NewBucket(impl, impl, configs.NewDefaultBucketConfig())
	}
}

func TestTokenAcquisition(t *testing.T) {
	for fqn, bucket := range testBuckets {
		fmt.Println("Testing ", fqn)

		// Clear any stale state
		bucket.Take(1, 0)

		wait := bucket.Take(1, 0)
		if wait != 0 {
			t.Fatalf("Expecting 0 wait. Was %v", wait)
		}

		// Consume all tokens. This should work too.
		if fqn == "redis:redis" {
			wait = bucket.Take(100, 0)
		} else {
			wait = bucket.Take(98, 0)
		}

		if wait != 0 {
			t.Fatalf("Expecting 0 wait. Was %v", wait)
		}

		// Should have no more left. Should have to wait.
		wait = bucket.Take(10, 0)
		if wait < 1 {
			t.Fatalf("Expecting positive wait time. Was %v", wait)
		}

		// If we don't want to wait...
		wait = bucket.Take(10, time.Nanosecond)
		if wait > -1 {
			t.Fatalf("Expecting negative wait time. Was %v", wait)
		}
	}
}
