// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package buckets defines interfaces for abstractions of token buckets.
package buckets

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
	"time"
	"strconv"
)

// Mock objects
type mockBucket struct {
	namespace, bucketName string
	dyn                   bool
}

func (b *mockBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	return 0
}
func (b *mockBucket) Config() *configs.BucketConfig {
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

func (bf mockBucketFactory) Init(cfg *configs.ServiceConfig) {}
func (bf mockBucketFactory) NewBucket(namespace string, bucketName string, cfg *configs.BucketConfig, dyn bool) Bucket {
	return &mockBucket{namespace: namespace, bucketName: bucketName, dyn: dyn}
}

var cfg = func() *configs.ServiceConfig {
	c := configs.NewDefaultServiceConfig()
	c.GlobalDefaultBucket = configs.NewDefaultBucketConfig()
	c.Namespaces["x"] = configs.NewDefaultNamespaceConfig()
	c.Namespaces["x"].DefaultBucket = configs.NewDefaultBucketConfig()
	c.Namespaces["x"].Buckets["a"] = configs.NewDefaultBucketConfig()

	c.Namespaces["y"] = configs.NewDefaultNamespaceConfig()
	c.Namespaces["y"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
	c.Namespaces["y"].Buckets["a"] = configs.NewDefaultBucketConfig()

	c.Namespaces["z"] = configs.NewDefaultNamespaceConfig()
	c.Namespaces["z"].MaxDynamicBuckets = 5
	c.Namespaces["z"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["a"] = configs.NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["b"] = configs.NewDefaultBucketConfig()
	c.Namespaces["z"].Buckets["c"] = configs.NewDefaultBucketConfig()
	return c
}()

var container = NewBucketContainer(cfg, &mockBucketFactory{})

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
