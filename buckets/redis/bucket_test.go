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
	"testing"
	"github.com/maniksurtani/quotaservice/buckets"
	"gopkg.in/redis.v3"
	"github.com/maniksurtani/quotaservice/configs"
	"os"
)

var cfg = configs.NewDefaultServiceConfig()
var factory buckets.BucketFactory
var bucket *redisBucket

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	factory = NewBucketFactory(&redis.Options{Addr: "localhost:6379"})
	factory.Init(cfg)
	bucket = factory.NewBucket("redis", "redis", configs.NewDefaultBucketConfig()).(*redisBucket)
}

func TestScriptLoaded(t *testing.T) {
	if !checkScriptExists(bucket.factory.client, bucket.factory.scriptSHA) {
		t.Fatal("Script not loaded into Redis at start")
	}
}

func TestFailingRedisConn(t *testing.T) {
	w := bucket.Take(1, 0)
	if w != 0 {
		t.Fatalf("Should have not seen any wait time. Saw %v", w)
	}

	err := bucket.factory.client.Close()
	if err != nil {
		t.Fatal("Couldn't kill client.")
	}

	// Client should reconnect
	w = bucket.Take(1, 0)
	if w != 0 {
		t.Fatalf("Should have not seen any wait time. Saw %v", w)
	}
}
