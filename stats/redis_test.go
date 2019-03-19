// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package stats

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"gopkg.in/redis.v5"

	"github.com/square/quotaservice/events"
)

var namespace string

const (
	characterList      = "abcdefghijklmnopqrstuvwxyz0123456789"
	characterListLen   = len(characterList)
	randomNamespaceLen = 16
)

func randomNamespace() string {
	result := make([]byte, randomNamespaceLen)

	for i := 0; i < randomNamespaceLen; i++ {
		result[i] = characterList[rand.Intn(characterListLen)]
	}

	return string(result)
}

const (
	batchSubmitInterval = 2 * time.Millisecond
	waitForBatchSubmit = 10 * batchSubmitInterval
)

func setUp() Listener {
	rand.Seed(time.Now().UTC().UnixNano())
	return NewRedisStatsListener(&redis.Options{Addr: "localhost:6379"}, 128, batchSubmitInterval)
}

func teardown(listener Listener) {
	listener.(*redisListener).client.FlushDb()
}

func TestRedisHandleNewHitBucket(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	ev := events.NewTokensServedEvent(namespace, "new-hit-dyn", true, 1, 0)
	listener.HandleEvent(ev)
	time.Sleep(waitForBatchSubmit)
	scores := listener.Get(namespace, "new-hit-dyn")

	if scores.Hits != 1 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=1, Misses=0]", scores)
	}

	ev = events.NewTokensServedEvent(namespace, "nondyn", false, 1, 0)
	listener.HandleEvent(ev)
	time.Sleep(waitForBatchSubmit)
	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	time.Sleep(waitForBatchSubmit)
	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Nonexisting namespace was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	teardown(listener)
}

func TestRedisHandleNewMissBucket(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "new-miss-dyn", true))
	time.Sleep(waitForBatchSubmit)
	scores := listener.Get(namespace, "new-miss-dyn")

	if scores.Hits != 0 || scores.Misses != 1 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=1]", scores)
	}

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "nondyn", false))
	time.Sleep(waitForBatchSubmit)
	scores = listener.Get(namespace, "nondyn")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Non-dynamic bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	teardown(listener)
}

func TestRedisHandleIncrMissBucket(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "incr-miss", true))
	time.Sleep(waitForBatchSubmit)
	scores := listener.Get(namespace, "incr-miss")

	if scores.Hits != 0 || scores.Misses != 3 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=3]", scores)
	}

	teardown(listener)
}

func TestRedisHandleIncrHitBucket(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 1, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "incr-hit", true, 1, 0))
	time.Sleep(waitForBatchSubmit)
	scores := listener.Get(namespace, "incr-hit")

	if scores.Hits != 5 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=5, Misses=0]", scores)
	}

	teardown(listener)
}

func TestRedisHandleNonEvent(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewTimedOutEvent(namespace, "nonevent", true, 1))
	time.Sleep(waitForBatchSubmit)
	scores := listener.Get(namespace, "nonevent")

	if scores.Hits != 0 || scores.Misses != 0 {
		t.Fatalf("Bucket score was not accurate: %+v != [Hits=0, Misses=0]", scores)
	}

	teardown(listener)
}

func TestRedisTop10Hits(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-1", true, 3, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-2", true, 10, 0))
	listener.HandleEvent(events.NewTokensServedEvent(namespace, "hits-dyn-3", true, 1, 0))

	time.Sleep(waitForBatchSubmit)
	hits := listener.TopHits(namespace)
	correctHits := []*BucketScore{
		{Bucket: "hits-dyn-2", Score: 10},
		{Bucket: "hits-dyn-1", Score: 3},
		{Bucket: "hits-dyn-3", Score: 1}}

	if !reflect.DeepEqual(hits, correctHits) {
		t.Fatalf("Hits top10 is not correct %+v", hits)
	}

	teardown(listener)
}

func TestRedisTop10Misses(t *testing.T) {
	listener := setUp()

	namespace = randomNamespace()
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-1", true))

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-2", true))

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-3", true))
	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-3", true))

	time.Sleep(waitForBatchSubmit)
	misses := listener.TopMisses(namespace)
	correctMisses := []*BucketScore{
		{Bucket: "misses-dyn-2", Score: 3},
		{Bucket: "misses-dyn-3", Score: 2},
		{Bucket: "misses-dyn-1", Score: 1}}

	if !reflect.DeepEqual(misses, correctMisses) {
		t.Fatalf("Misses top10 is not correct %+v", misses)
	}

	teardown(listener)
}

func TestRedisBatching( t *testing.T) {
	setUp()
	listener := NewRedisStatsListener(&redis.Options{Addr: "localhost:6379"}, 128, 5 * time.Second)

	for i := 0; i < 127; i++ {
		listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-1", true))
	}

	misses := listener.TopMisses(namespace)
	if !reflect.DeepEqual(misses, []*BucketScore{}) {
		t.Fatalf("Misses top1 was not empty %+v", misses)
	}

	listener.HandleEvent(events.NewBucketMissedEvent(namespace, "misses-dyn-1", true))
	time.Sleep(waitForBatchSubmit)
	misses = listener.TopMisses(namespace)
	correctMisses := []*BucketScore{{Bucket: "misses-dyn-1", Score: 128}}

	if !reflect.DeepEqual(misses, correctMisses) {
		t.Fatalf("Misses top1 is not correct %+v", misses)
	}

	teardown(listener)
}
