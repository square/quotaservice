package buckets

import (
	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/config"
	"strconv"
	"testing"
	"time"
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
	cfg := config.NewDefaultServiceConfig()
	cfg.Namespaces["n"] = config.NewDefaultNamespaceConfig()
	cfg.Namespaces["n"].DynamicBucketTemplate = config.NewDefaultBucketConfig()
	// Times out every 250 millis.
	cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis = 250
	events := &quotaservice.MockEmitter{Events: make(chan quotaservice.Event, 100)}
	container := quotaservice.NewBucketContainer(cfg, factory, events)

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

	bucketNames := make([]string, 10)
	for i := 0; i < 10; i++ {
		bucketNames[i] = strconv.Itoa(i)
	}

	waitForGC(events.Events, "n", bucketNames)

	for _, bName := range bucketNames {
		// Check that the bucket has been GC'd
		if container.Exists("n", bName) {
			t.Fatalf("Bucket %v wasn't GC'd when it should have on impl %v", bName, impl)
		}
	}
}

func waitForGC(events chan quotaservice.Event, namespace string, buckets []string) {
	bucketMap := make(map[string]bool)
	for _, b := range buckets {
		bucketMap[b] = true
	}

	for e := range events {
		if e.EventType() == quotaservice.EVENT_BUCKET_REMOVED && e.Namespace() == namespace {
			bucketMap[e.BucketName()] = false
		}

		// Scan bucketMap
		unseenBuckets := false
		for _, waiting := range bucketMap {
			unseenBuckets = unseenBuckets || waiting
		}

		if !unseenBuckets {
			return
		}
	}
}
