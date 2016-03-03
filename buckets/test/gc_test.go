// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements bucket tests
package test

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/logging"
	"strconv"
	"time"
)

func TestGC(t *testing.T) {
	for impl, factory := range factories {
		logging.Printf("Testing %v", impl)
		cfg := configs.NewDefaultServiceConfig()
		cfg.Namespaces["n"] = configs.NewDefaultNamespaceConfig()
		cfg.Namespaces["n"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
		// Times out every 5 seconds
		cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis = 5000
		container := buckets.NewBucketContainer(cfg, factory)

		// No GC should happen here as long as we are in use.
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				bName := strconv.Itoa(j)
				b, _ := container.FindBucket("n", bName)
				if b == nil {
					t.Fatalf("Failed looking for bucket %v on impl %v", bName, impl)
				}

				// Check that the bucket hasn't been GC'd
				if !container.Exists("n", bName) {
					t.Fatalf("Bucket %v was GC'd when it shouldn't have on impl %v", bName, impl)
				}
			}
		}

		// Time out.
		time.Sleep(time.Duration(cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis) * time.Millisecond * 4)

		// GC should happen here after sleep.
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				bName := strconv.Itoa(j)
				// Check that the bucket has been GC'd
				if container.Exists("n", bName) {
					t.Fatalf("Bucket %v wasn't GC'd when it should have on impl %v", bName, impl)
				}
			}
		}
	}
}
