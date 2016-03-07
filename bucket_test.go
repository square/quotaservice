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
	retval                time.Duration
	active                bool
	namespace, bucketName string
	dyn                   bool
	cfg                   *BucketConfig
}

func (b *mockBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	return b.retval
}
func (b *mockBucket) Config() *BucketConfig {
	return b.cfg
}
func (b *mockBucket) ActivityDetected() bool {
	return b.active
}
func (b *mockBucket) ReportActivity() {}
func (b *mockBucket) Dynamic() bool {
	return b.dyn
}
func (b *mockBucket) Destroy() {}

type mockBucketFactory struct {
	buckets map[string]*mockBucket
}

func (bf *mockBucketFactory) SetRetval(namespace, name string, d time.Duration) {
	bf.buckets[FullyQualifiedName(namespace, name)].retval = d
}

func (bf *mockBucketFactory) MarkForGC(namespace, name string) {
	bf.buckets[FullyQualifiedName(namespace, name)].active = false
}

func (bf *mockBucketFactory) Init(cfg *ServiceConfig) {}
func (bf *mockBucketFactory) NewBucket(namespace string, bucketName string, cfg *BucketConfig, dyn bool) Bucket {
	b := &mockBucket{0, true, namespace, bucketName, dyn, cfg}
	if bf.buckets == nil {
		bf.buckets = make(map[string]*mockBucket)
	}

	bf.buckets[FullyQualifiedName(namespace, bucketName)] = b
	return b
}

var cfg = func() *ServiceConfig {
	c := NewDefaultServiceConfig()
	c.GlobalDefaultBucket = NewDefaultBucketConfig()

	// Namespace "x"
	ns := NewDefaultNamespaceConfig()
	ns.DefaultBucket = NewDefaultBucketConfig()
	ns.AddBucket("a", NewDefaultBucketConfig())
	c.AddNamespace("x", ns)

	// Namespace "y"
	ns = NewDefaultNamespaceConfig()
	ns.DynamicBucketTemplate = NewDefaultBucketConfig()
	ns.AddBucket("a", NewDefaultBucketConfig())
	c.AddNamespace("y", ns)

	// Namespace "z"
	ns = NewDefaultNamespaceConfig()
	ns.DynamicBucketTemplate = NewDefaultBucketConfig()
	ns.MaxDynamicBuckets = 5
	ns.
	AddBucket("a", NewDefaultBucketConfig()).
	AddBucket("b", NewDefaultBucketConfig()).
	AddBucket("c", NewDefaultBucketConfig())
	c.AddNamespace("z", ns)

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
