// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE
package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/square/quotaservice/buckets"
	"github.com/square/quotaservice/config"
	quotaservice_configs "github.com/square/quotaservice/protos/config"
)

const (
	dynMaxDebtMillis = 777
)

var cfg *quotaservice_configs.ServiceConfig
var factory *bucketFactory
var bucket *staticBucket
var dynMaxDebtNanos = fmt.Sprintf("%v", dynMaxDebtMillis*1000000)

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	dynNs := config.NewDefaultNamespaceConfig("dynNs")
	dynTpl := config.NewDefaultBucketConfig(config.DynamicBucketTemplateName)
	dynTpl.MaxDebtMillis = dynMaxDebtMillis
	config.SetDynamicBucketTemplate(dynNs, dynTpl)

	cfg = config.NewDefaultServiceConfig()
	config.AddNamespace(cfg, dynNs)

	factory = NewBucketFactory(&redis.Options{Addr: "localhost:6379"}, 2, 0).(*bucketFactory)
	factory.Init(cfg)
	bucket = factory.NewBucket("redis", "redis", config.NewDefaultBucketConfig(""), false).(*staticBucket)
}

func TestScriptLoaded(t *testing.T) {
	bucket.Take(context.Background(), 1, 50*time.Millisecond)

	exists := bucket.factory.script.Exists(context.TODO(), bucket.factory.client)
	if err := exists.Err(); err != nil {
		t.Fatal("Error checking if script was loaded into Redis: ", err)
	} else if !exists.Val()[0] {
		t.Fatal("Script not loaded into Redis")
	}
}

func TestFailingRedisConn(t *testing.T) {
	w, s, err := bucket.Take(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if w < 0 {
		t.Fatalf("Should have not seen negative wait time. Saw %v", w)
	}
	if !s {
		t.Fatalf("Success should be true.")
	}

	err = bucket.factory.client.Close()
	if err != nil {
		t.Fatal("Couldn't kill client.")
	}

	// Client should fail to Take(). This should start the reconnect handler
	w, s, err = bucket.Take(context.Background(), 1, 0)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if w < 0 {
		t.Fatalf("Should have not seen negative wait time. Saw %v", w)
	}
	if s {
		t.Fatalf("Success should be false.")
	}

	for numTimeWaited := bucket.factory.connectionRetries; bucket.factory.getNumTimesConnResolved() == 0 && numTimeWaited > 0; numTimeWaited-- {
		time.Sleep(5 * time.Second)
	}

	// Client should reconnect
	w, s, err = bucket.Take(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
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

func TestRefCountsForDynamic(t *testing.T) {
	assertNoSharedAttribs(t)

	// Create a dynamic bucket - the cast will ensure it is the right type
	b1 := factory.NewBucket("dynNs", "b1", cfg.Namespaces["dynNs"].DynamicBucketTemplate, true).(*dynamicBucket)
	assertRefCounts("dynNs", 1, t)
	if b1.maxDebtNanos != dynMaxDebtNanos {
		t.Fatalf("Expected maxDebtNanos on dynamic bucket to be %v but was %v", dynMaxDebtNanos, b1.maxDebtNanos)
	}

	b2 := factory.NewBucket("dynNs", "b2", cfg.Namespaces["dynNs"].DynamicBucketTemplate, true).(*dynamicBucket)
	assertRefCounts("dynNs", 2, t)

	// Check that b1 and b2 point to the same shared attributes
	if b1.abstractBucket.configAttributes != b2.abstractBucket.configAttributes {
		t.Fatalf("b1 and b2 point to different configAttributes. b1 points to %p and b2 points to %p", b1.abstractBucket.configAttributes, b2.abstractBucket.configAttributes)
	}

	b2.Destroy()
	assertRefCounts("dynNs", 1, t)

	b3 := factory.NewBucket("dynNs", "b3", cfg.Namespaces["dynNs"].DynamicBucketTemplate, true).(*dynamicBucket)
	// Check that b1 and b3 point to the same shared attributes
	if b1.abstractBucket.configAttributes != b3.abstractBucket.configAttributes {
		t.Fatalf("b1 and b3 point to different configAttributes. b1 points to %p and b3 points to %p", b1.abstractBucket.configAttributes, b3.abstractBucket.configAttributes)
	}

	assertRefCounts("dynNs", 2, t)

	b1.Destroy()
	b3.Destroy()

	assertNoSharedAttribs(t)
}

func assertNoSharedAttribs(t *testing.T) {
	t.Helper()

	if len(factory.sharedAttributes) != 0 {
		t.Fatalf("Expected no shared attributes. Found %+v and refcounts %+v", factory.sharedAttributes, factory.refcounts)
	}

	if len(factory.refcounts) != 0 {
		t.Fatalf("Expected no references to shared attributes. Found %+v", factory.refcounts)
	}
}

func assertRefCounts(ns string, expected int, t *testing.T) {
	t.Helper()

	if _, exists := factory.sharedAttributes[ns]; !exists {
		t.Fatalf("Expected shared attributes for namespace %v but found %+v", ns, factory.sharedAttributes)
	}

	if counts, exists := factory.refcounts[ns]; !exists {
		t.Fatalf("Expected ref counts for namespace %v but found %+v", ns, factory.refcounts)
	} else {
		if counts != expected {
			t.Fatalf("Expected ref counts for namespace %v to be %v but found %v", ns, expected, counts)
		}
	}
}
