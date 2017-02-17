// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"github.com/maniksurtani/quotaservice/config"
	pbc "github.com/maniksurtani/quotaservice/protos/config"
	"testing"
	"time"
)

func TestNotReapable(t *testing.T) {
	_, bc, r := reaperSetup()
	tb := &MockBucket{}
	b, _ := r.applyWatch(tb, "x", "y", config.NewDefaultBucketConfig("y"))

	if _, ok := b.(*MockBucket); !ok {
		t.Fatalf("OK is false; b is of type %T", b)
	}

	reaperTeardown(bc)
}

func TestWatcher(t *testing.T) {
	_, bc, _ := reaperSetup()
	b, w := createTestReapableBucket(10000, bc)

	if w.ns != "x" {
		t.Fatal("Wrong namespace")
	}

	if w.bucketName != "y" {
		t.Fatal("Wrong bucket name")
	}

	if w.activities == nil {
		t.Fatal("Activity monitor is nil")
	}

	b.ReportActivity()

	if !w.activityDetected() {
		t.Fatal("Didn't detect activity")
	}

	if w.activityDetected() {
		t.Fatal("Detected phantom activity")
	}

	// Don't test watcher.tooIdle() with the reaper goroutine running, or gotest will warn of races.
	reaperTeardown(bc)
}

func TestIdleAndActivity(t *testing.T) {
	_, bc, _ := reaperSetup()
	// Create watcher but don't let the reaper's goroutine use it.
	b, ac := createTestReapableBucketNoWatch()
	w := createWatcher("x", "y", 1000*time.Millisecond, ac)

	now := time.Now()
	b.ReportActivity()
	if w.tooIdle(now) {
		t.Fatal("Too idle too early")
	}

	// This one won't consume the activity.
	if w.tooIdle(now) {
		t.Fatal("Too idle too early")
	}

	if !w.tooIdle(now.Add(time.Hour)) {
		t.Fatal("Should be too idle!")
	}

	// Should be idempotent
	if !w.tooIdle(now.Add(2 * time.Hour)) {
		t.Fatal("Should be too idle!")
	}

	// Activity should clear it
	b.ReportActivity()
	if w.tooIdle(now.Add(3 * time.Hour)) {
		t.Fatal("Should have detected activity!")
	}

	reaperTeardown(bc)
}

func createTestReapableBucket(maxIdle int64, bc *bucketContainer) (*reapableBucket, *watcher) {
	tb := &MockBucket{}
	c := config.NewDefaultBucketConfig("y")
	c.MaxIdleMillis = maxIdle
	b, w := bc.r.applyWatch(tb, "x", "y", c)
	return b.(*reapableBucket), w
}

func createTestReapableBucketNoWatch() (*reapableBucket, <-chan struct{}) {
	tb := &MockBucket{}
	ch := make(chan struct{}, 1)
	return &reapableBucket{Bucket: tb, activities: ch}, ch
}

func reaperSetup() (*pbc.ServiceConfig, *bucketContainer, *reaper) {
	cfg := config.NewDefaultServiceConfig()
	bc, _, _ := NewBucketContainerWithMocks(cfg)
	return cfg, bc, bc.r
}

func reaperTeardown(bc *bucketContainer) {
	bc.Stop()
}
