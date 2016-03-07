// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package buckets defines interfaces for abstractions of token buckets.
package quotaservice

import (
	"strconv"
	"testing"
	"time"
)

// Mock objects
type mockBucket struct {
	namespace, bucketName string
	dyn                   bool
}

func (b *mockBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	return 0
}
func (b *mockBucket) Config() *BucketConfig {
	return nil
}
func (b *mockBucket) ActivityDetected() bool {
	return true
}
func (b *mockBucket) ReportActivity() {}
func (b *mockBucket) Dynamic() bool {
	return b.dyn
}
func (b *mockBucket) Destroy() {}

type mockBucketFactory struct{}

func (bf mockBucketFactory) Init(cfg *ServiceConfig) {}
func (bf mockBucketFactory) NewBucket(namespace string, bucketName string, cfg *BucketConfig, dyn bool) Bucket {
	return &mockBucket{namespace: namespace, bucketName: bucketName, dyn: dyn}
}

var cfg = func() *ServiceConfig {
	c := NewDefaultServiceConfig()
	c.GlobalDefaultBucket = NewDefaultBucketConfig()
	c.Namespaces["x"] = NewDefaultNamespaceConfig()
	c.Namespaces["x"].DefaultBucket = NewDefaultBucketConfig()
	c.Namespaces["x"].Buckets["a"] = NewDefaultBucketConfig()

	c.Namespaces["y"] = NewDefaultNamespaceConfig()
	c.Namespaces["y"].DynamicBucketTemplate = NewDefaultBucketConfig()
	c.Namespaces["y"].Buckets["a"] = NewDefaultBucketConfig()

	c.Namespaces["z"] = NewDefaultNamespaceConfig()
	c.Namespaces["z"].MaxDynamicBuckets = 5
	c.Namespaces["z"].DynamicBucketTemplate = NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["a"] = NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["b"] = NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["c"] = NewDefaultBucketConfig()
	return c
}()

var container = NewBucketContainer(cfg, &mockBucketFactory{}, &MockEmitter{})

func TestFallbackToGlobalDefaultBucket(t *testing.T) {
	b, _ := container.FindBucket("nonexistent_namespace", "nonexistent_bucket")

	if b == nil {
		t.Fatal("Should fall back to default bucket.")
	}

	if b != container.defaultBucket {
		t.Fatal("Should fall back to default bucket.")
	}
}

func TestFallbackToDefaultBucket(t *testing.T) {
	b, _ := container.FindBucket("x", "nonexistent_bucket")
	if b == nil {
		t.Fatal("Should fall back to default bucket.")
	}

	if b != container.namespaces["x"].defaultBucket {
		t.Fatal("Should fall back to default bucket.")
	}
}

func TestDynamicBucket(t *testing.T) {
	b, _ := container.FindBucket("y", "new")
	if b == nil {
		t.Fatal("Should create new bucket.")
	}

	if b != container.namespaces["y"].buckets["new"] {
		t.Fatal("Should create new bucket.")
	}
}

func TestBucketNamespaces(t *testing.T) {
	bx, _ := container.FindBucket("x", "a")
	if bx == nil {
		t.Fatal("Should create new bucket.")
	}

	if bx != container.namespaces["x"].buckets["a"] {
		t.Fatal("Should create new bucket.")
	}

	by, _ := container.FindBucket("y", "a")
	if by == nil {
		t.Fatal("Should create new bucket.")
	}

	if by != container.namespaces["y"].buckets["a"] {
		t.Fatal("Should create new bucket.")
	}

	if by == bx {
		t.Fatal("Buckets in different namespaces should be different instances.")
	}
}

func TestMaxDynamic(t *testing.T) {
	c := container.countDynamicBuckets("z")
	if c != 0 {
		t.Fatalf("Should have 0 dynamic buckets. Instead was %v", c)
	}

	for i := 0; i < 5; i++ {
		container.createNewNamedBucket("z", strconv.Itoa(i), container.namespaces["z"])
	}

	c = container.countDynamicBuckets("z")
	if c != 5 {
		t.Fatalf("Should have 5 dynamic buckets. Instead was %v", c)
	}

	b := container.createNewNamedBucket("z", "should_fail", container.namespaces["z"])
	if b != nil {
		t.Fatal("Should not have created dynamic bucket z:should_fail")
	}
}
