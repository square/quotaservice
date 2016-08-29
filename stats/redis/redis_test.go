// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package redis

import (
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"gopkg.in/redis.v3"

	"github.com/maniksurtani/quotaservice/events"
	"github.com/maniksurtani/quotaservice/stats"
)

var listener stats.Listener
var namespace string

const (
	characterList      = "abcdefghijklmnopqrstuvwxyz0123456789"
	characterListLen   = len(characterList)
	randomNamespaceLen = 16
)

func TestMain(m *testing.M) {
	setUp()
	os.Exit(m.Run())
}

func randomNamespace() string {
	result := make([]byte, randomNamespaceLen)

	for i := 0; i < randomNamespaceLen; i++ {
		result[i] = characterList[rand.Intn(characterListLen)]
	}

	return string(result)
}

func setUp() {
	rand.Seed(time.Now().UTC().UnixNano())
	listener = stats.NewRedisStatsListener(&redis.Options{Addr: "localhost:6379"})
}

func TestHandleNewHitBucket(t *testing.T) {
	namespace = randomNamespace()
	ev := events.NewTokensServedEvent(namespace, "new-hit-dyn", true, 1, 0)
	listener.HandleEvent(ev)
	scores := listener.Get(namespace, "new-hit-dyn")

	if scores.Hits != 1 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=1, Misses=0]", scores)
	}

	ev = events.NewTokensServedEvent(namespace, "nondyn", false, 1, 0)
	listener.HandleEvent(ev)
	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Nonexisting namespace was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestHandleNewMissBucket(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "new-miss-dyn", true))
	scores := listener.Get(namespace, "new-miss-dyn")

	if scores.Hits != 0 || scores.Misses != 1 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=1]", scores)
	}

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "nondyn", false))
	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestHandleIncrMissBucket(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	scores := listener.Get(namespace, "incr-miss")

	if scores.Hits != 0 || scores.Misses != 3 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=3]", scores)
	}
}

func TestHandleIncrHitBucket(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 1, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 1, 0))
	scores := listener.Get(namespace, "incr-hit")

	if scores.Hits != 5 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=5, Misses=0]", scores)
	}
}

func TestHandleNonEvent(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewTimedOutEvent(namespace, "nonevent", true, 1))
	scores := listener.Get(namespace, "nonevent")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}
}

func TestTop10Hits(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-1", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-2", true, 10, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-3", true, 1, 0))

	hits := listener.TopHits(namespace)
	correctHits := []*stats.BucketScore{
		&stats.BucketScore{"hits-dyn-2", 10},
		&stats.BucketScore{"hits-dyn-1", 3},
		&stats.BucketScore{"hits-dyn-3", 1}}

	if !reflect.DeepEqual(hits, correctHits) {
		t.Fatalf("Hits top10 is not correct %+v", hits)
	}
}

func TestTop10Misses(t *testing.T) {
	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-1", true))

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-3", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-3", true))

	misses := listener.TopMisses(namespace)
	correctMisses := []*stats.BucketScore{
		&stats.BucketScore{"misses-dyn-2", 3},
		&stats.BucketScore{"misses-dyn-3", 2},
		&stats.BucketScore{"misses-dyn-1", 1}}

	if !reflect.DeepEqual(misses, correctMisses) {
		t.Fatalf("Misses top10 is not correct %+v", misses)
	}
}
