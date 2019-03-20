package buckets

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/square/quotaservice"
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/logging"
	"github.com/square/quotaservice/test/helpers"
)

func TestTokenAcquisition(t *testing.T, bucket quotaservice.Bucket) {
	// Clear any stale state
	bucket.Take(context.Background(), 1, 0)

	wait, s := bucket.Take(context.Background(), 1, 0)
	if wait != 0 {
		t.Fatalf("Expecting 0 wait. Was %v", wait)
	}
	if !s {
		t.Fatal("Expecting success to be true.")
	}

	// Consume all tokens. This should work too.
	wait, s = bucket.Take(context.Background(), 100, 0)

	if wait != 0 {
		t.Fatalf("Expecting 0 wait. Was %v", wait)
	}
	if !s {
		t.Fatal("Expecting success to be true.")
	}

	// Should have no more left. Should have to wait.
	wait, s = bucket.Take(context.Background(), 10, 10*time.Second)
	if wait < 1 {
		t.Fatalf("Expecting positive wait time. Was %v", wait)
	}
	if !s {
		t.Fatal("Expecting success to be true.")
	}

	// If we don't want to wait...
	wait, s = bucket.Take(context.Background(), 10, 0)
	if wait != 0 {
		t.Fatalf("Expecting 0 wait time. Was %v", wait)
	}
	if s {
		t.Fatal("Expecting success to be false.")
	}
}

func TestGC(t *testing.T, factory quotaservice.BucketFactory, impl string) {
	cfg := config.NewDefaultServiceConfig()
	nsCfg := config.NewDefaultNamespaceConfig("n")
	tpl := config.NewDefaultBucketConfig("")
	// Times out every 250 millis.
	tpl.MaxIdleMillis = 250
	config.SetDynamicBucketTemplate(nsCfg, tpl)
	helpers.CheckError(t, config.AddNamespace(cfg, nsCfg))

	eventsEmitter := &quotaservice.MockEmitter{Events: make(chan events.Event, 100)}
	container := quotaservice.NewBucketContainer(factory, eventsEmitter, quotaservice.NewReaperConfigForTests())
	container.Init(cfg)

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

	waitForGC(eventsEmitter.Events, "n", bucketNames)

	for _, bName := range bucketNames {
		// Check that the bucket has been GC'd
		if container.Exists("n", bName) {
			t.Fatalf("Bucket %v wasn't GC'd when it should have on impl %v", bName, impl)
		}
	}
}

func waitForGC(eventsChan <-chan events.Event, namespace string, buckets []string) {
	logging.Info("Waiting for GC")
	bucketMap := make(map[string]bool)
	for _, b := range buckets {
		bucketMap[b] = true
	}

	for e := range eventsChan {
		if e.EventType() == events.EVENT_BUCKET_REMOVED && e.Namespace() == namespace {
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
