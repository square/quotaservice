// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package stats

import (
	"reflect"
	"testing"

	"github.com/square/quotaservice/events"
)

var listener Listener

func TestMemoryHandleNewHitBucket(t *testing.T) {
	listener = NewMemoryStatsListener()
	ev := events.NewTokensServedEvent("test", "dyn", true, 1, 0)
	listener.HandleEvent(ev)
	scores := listener.Get("test", "dyn")

	if scores.Hits != 1 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=1, Misses=0]", scores)
	}

	ev = events.NewTokensServedEvent("test", "nondyn", false, 1, 0)
	listener.HandleEvent(ev)
	scores = listener.Get("test", "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	scores = listener.Get("nontest", "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Nonexisting namespace was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestMemoryHandleNewMissBucket(t *testing.T) {
	listener = NewMemoryStatsListener()
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn", true))
	scores := listener.Get("test", "dyn")

	if scores.Hits != 0 || scores.Misses != 1 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=1]", scores)
	}

	listener.HandleEvent(events.NewBucketMissedEvent("test", "nondyn", false))
	scores = listener.Get("test", "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestMemoryHandleIncrMissBucket(t *testing.T) {
	listener = NewMemoryStatsListener()
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn", true))
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn", true))
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn", true))
	scores := listener.Get("test", "dyn")

	if scores.Hits != 0 || scores.Misses != 3 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=3]", scores)
	}
}

func TestMemoryHandleIncrHitBucket(t *testing.T) {
	listener = NewMemoryStatsListener()
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn", true, 1, 0))
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn", true, 1, 0))
	scores := listener.Get("test", "dyn")

	if scores.Hits != 5 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=5, Misses=0]", scores)
	}
}

func TestMemoryHandleNonEvent(t *testing.T) {
	listener = NewMemoryStatsListener()
	listener.HandleEvent(events.NewTimedOutEvent("test", "dyn", true, 1))
	scores := listener.Get("test", "dyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestMemoryTop10Hits(t *testing.T) {
	listener = NewMemoryStatsListener()
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn-1", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn-2", true, 10, 0))
	listener.HandleEvent(events.NewTokensServedEvent("test", "dyn-3", true, 1, 0))

	hits := listener.TopHits("test")
	correctHits := []*BucketScore{
		{Bucket: "dyn-2", Score: 10},
		{Bucket: "dyn-1", Score: 3},
		{Bucket: "dyn-3", Score: 1}}

	if !reflect.DeepEqual(hits, correctHits) {
		t.Fatalf("Hits top10 is not correct %+v", hits)
	}
}

func TestMemoryTop10Misses(t *testing.T) {
	listener = NewMemoryStatsListener()

	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-1", true))

	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-2", true))

	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-3", true))
	listener.HandleEvent(events.NewBucketMissedEvent("test", "dyn-3", true))

	misses := listener.TopMisses("test")
	correctMisses := []*BucketScore{
		{Bucket: "dyn-2", Score: 3},
		{Bucket: "dyn-3", Score: 2},
		{Bucket: "dyn-1", Score: 1}}

	if !reflect.DeepEqual(misses, correctMisses) {
		t.Fatalf("Misses top10 is not correct %+v", misses)
	}
}
