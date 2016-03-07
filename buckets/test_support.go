package buckets

import (
	"testing"
	"github.com/maniksurtani/quotaservice"
	"time"
	"strconv"
)

func TestTokenAcquisition(t *testing.T, bucket quotaservice.Bucket) {
	// Clear any stale state
	bucket.Take(1, 0)

	wait := bucket.Take(1, 0)
	if wait != 0 {
		t.Fatalf("Expecting 0 wait. Was %v", wait)
	}

	// Consume all tokens. This should work too.
	wait = bucket.Take(100, 0)

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

func TestGC(t *testing.T, factory quotaservice.BucketFactory, impl string) {
	cfg := quotaservice.NewDefaultServiceConfig()
	cfg.Namespaces["n"] = quotaservice.NewDefaultNamespaceConfig()
	cfg.Namespaces["n"].DynamicBucketTemplate = quotaservice.NewDefaultBucketConfig()
	// Times out every 5 seconds
	cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis = 5000
	container := quotaservice.NewBucketContainer(cfg, factory, &quotaservice.MockEmitter{})

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
