// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package redis

import (
	"testing"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/config"
	"gopkg.in/redis.v3"
	"os"
)

var cfg = config.NewDefaultServiceConfig()
var factory quotaservice.BucketFactory
var bucket *redisBucket

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	factory = NewBucketFactory(&redis.Options{Addr: "localhost:6379"}, 2)
	factory.Init(cfg)
	bucket = factory.NewBucket("redis", "redis", config.NewDefaultBucketConfig(""), false).(*redisBucket)
}

func TestScriptLoaded(t *testing.T) {
	if !checkScriptExists(bucket.factory.client, bucket.factory.scriptSHA) {
		t.Fatal("Script not loaded into Redis at start")
	}
}

func TestFailingRedisConn(t *testing.T) {
	w, s := bucket.Take(1, 0)

	if w < 0 {
		t.Fatalf("Should have not seen negative wait time. Saw %v", w)
	}
	if !s {
		t.Fatalf("Success should be true.")
	}

	err := bucket.factory.client.Close()
	if err != nil {
		t.Fatal("Couldn't kill client.")
	}

	// Client should reconnect
	w, s = bucket.Take(1, 0)
	if w < 0 {
		t.Fatalf("Should have not seen negative wait time. Saw %v", w)
	}
	if !s {
		t.Fatalf("Success should be true.")
	}
}

func TestTokenAcquisition(t *testing.T) {
	buckets.TestTokenAcquisition(t, bucket)
}

func TestGC(t *testing.T) {
	buckets.TestGC(t, factory, "redis")
}
