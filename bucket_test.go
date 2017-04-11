// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package buckets defines interfaces for abstractions of token buckets.
package quotaservice

import (
	"strconv"
	"testing"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/test/helpers"

	"runtime"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

var cfg = func() *pbconfig.ServiceConfig {
	c := config.NewDefaultServiceConfig()
	c.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)

	// Namespace "x"
	ns := config.NewDefaultNamespaceConfig("x")
	ns.DefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("a")))
	helpers.PanicError(config.AddNamespace(c, ns))

	// Namespace "y"
	ns = config.NewDefaultNamespaceConfig("y")
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig(config.DefaultBucketName)
	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("y")))
	helpers.PanicError(config.AddNamespace(c, ns))

	// Namespace "z"
	ns = config.NewDefaultNamespaceConfig("z")
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig(config.DefaultBucketName)
	ns.MaxDynamicBuckets = 5
	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("a")))
	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("b")))
	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("c")))
	helpers.PanicError(config.AddNamespace(c, ns))

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
	initGoroutineCount := runtime.NumGoroutine()
	b, _ := container.FindBucket("y", "new")
	if b == nil {
		t.Fatal("Should create new bucket.")
	}

	if b != container.namespaces["y"].buckets["new"] {
		t.Fatal("Should create new bucket.")
	}

	b2, _ := container.FindBucket("y", "new")
	if b == nil {
		t.Fatal("Should return a bucket.")
	}

	if b != b2 {
		t.Fatal("Should not create a new bucket.")
	}

	numGoroutines := runtime.NumGoroutine()
	if numGoroutines != initGoroutineCount {
		t.Fatalf("Expected no more additional goroutines to be created, but was %v.", numGoroutines-initGoroutineCount)
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
