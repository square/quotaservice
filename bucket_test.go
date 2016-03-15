// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package buckets defines interfaces for abstractions of token buckets.
package quotaservice

import (
	"strconv"
	"testing"
)

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

var container, _, _ = NewBucketContainerWithMocks(cfg)

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

func TestDelete(t *testing.T) {
	if !container.Exists("x", "a") {
		t.Fatal("x:a should exist")
	}

	container.deleteBucket("x", "a")

	if container.Exists("x", "a") {
		t.Fatal("x:a should not exist")
	}

	if container.defaultBucket == nil {
		t.Fatal("Should have a global default bucket.")
	}

	container.deleteBucket(globalNamespace, defaultBucketName)

	if container.defaultBucket != nil {
		t.Fatal("Should delete global default bucket.")
	}

	if container.namespaces["x"].defaultBucket == nil {
		t.Fatal("Default bucket on x should exist")
	}

	container.deleteBucket("x", defaultBucketName)

	if container.namespaces["x"].defaultBucket != nil {
		t.Fatal("Default bucket on x should not exist")
	}

	if container.namespaces["y"] == nil {
		t.Fatal("Namespace y should exist")
	}

	container.deleteNamespace("y")

	if container.namespaces["y"] != nil {
		t.Fatal("Namespace y should not exist")
	}
}
