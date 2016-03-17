// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"github.com/maniksurtani/quotaservice/config"
	"os"
	"testing"
	"time"
)

var s Server
var qs QuotaService
var events chan Event
var mbf *MockBucketFactory

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	cfg := config.NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = config.NewDefaultBucketConfig()

	// Namespace "dyn"
	ns := config.NewDefaultNamespaceConfig()
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig()
	ns.DynamicBucketTemplate.MaxTokensPerRequest = 5
	ns.DynamicBucketTemplate.MaxIdleMillis = 100
	ns.MaxDynamicBuckets = 3
	cfg.AddNamespace("dyn", ns)

	// Namespace "nodyn"
	ns = config.NewDefaultNamespaceConfig()
	b := config.NewDefaultBucketConfig()
	b.MaxTokensPerRequest = 10
	ns.AddBucket("b", b)
	cfg.AddNamespace("nodyn", ns)

	mbf = &MockBucketFactory{}
	me := &MockEndpoint{}
	s = New(cfg, mbf, me)
	events = make(chan Event, 100)
	s.SetListener(func(e Event) {
		events <- e
	}, 100)
	s.Start()
	qs = me.QuotaService
	// New buckets would have been created. Clear all notifications.
	clearEvents(1)
}

func TestTokens(t *testing.T) {
	qs.Allow("nodyn", "b", 1, 0)
	checkEvent("nodyn", "b", false, EVENT_TOKENS_SERVED, 1, 0, <-events, t)
}

func TestTooManyTokens(t *testing.T) {
	qs.Allow("nodyn", "b", 100, 0)
	checkEvent("nodyn", "b", false, EVENT_TOO_MANY_TOKENS_REQUESTED, 100, 0, <-events, t)
}

func TestTimeout(t *testing.T) {
	mbf.SetWaitTime("nodyn", "b", 2*time.Minute)
	qs.Allow("nodyn", "b", 1, 1)
	checkEvent("nodyn", "b", false, EVENT_TIMEOUT_SERVING_TOKENS, 1, 0, <-events, t)
	mbf.SetWaitTime("nodyn", "b", 0)
}

func TestWithWait(t *testing.T) {
	mbf.SetWaitTime("nodyn", "b", 2*time.Nanosecond)
	qs.Allow("nodyn", "b", 1, 0)
	checkEvent("nodyn", "b", false, EVENT_TOKENS_SERVED, 1, 2*time.Nanosecond, <-events, t)
	mbf.SetWaitTime("nodyn", "b", 0)
}

func TestNoSuchBucket(t *testing.T) {
	qs.Allow("nodyn", "x", 1, 0)
	checkEvent("nodyn", "x", false, EVENT_BUCKET_MISS, 0, 0, <-events, t)
}

func TestNewDynBucket(t *testing.T) {
	qs.Allow("dyn", "b", 1, 0)
	checkEvent("dyn", "b", true, EVENT_BUCKET_CREATED, 0, 0, <-events, t)
	checkEvent("dyn", "b", true, EVENT_TOKENS_SERVED, 1, 0, <-events, t)
}

func TestTooManyDynBuckets(t *testing.T) {
	qs.Allow("dyn", "c", 1, 0)
	qs.Allow("dyn", "d", 1, 0)
	clearEvents(4)

	qs.Allow("dyn", "e", 1, 0)
	checkEvent("dyn", "e", true, EVENT_BUCKET_MISS, 0, 0, <-events, t)
}

func TestBucketRemoval(t *testing.T) {
	// GC thread should run every 100ms for this namespace. Make sure it runs at least once.
	time.Sleep(300 * time.Millisecond)

	mbf.SetActive("dyn", "b", false)
	checkEvent("dyn", "b", true, EVENT_BUCKET_REMOVED, 0, 0, <-events, t)
	mbf.SetActive("dyn", "c", false)
	checkEvent("dyn", "c", true, EVENT_BUCKET_REMOVED, 0, 0, <-events, t)
	mbf.SetActive("dyn", "d", false)
	checkEvent("dyn", "d", true, EVENT_BUCKET_REMOVED, 0, 0, <-events, t)
}

func checkEvent(namespace, name string, dyn bool, eventType EventType, tokens int64, waitTime time.Duration, actual Event, t *testing.T) {
	if actual == nil {
		t.Fatalf("Expecting event; was nil.")
	}

	if actual.Namespace() != namespace {
		t.Fatalf("Event should have namespace '%v'. Was '%v'.", namespace, actual.Namespace())
	}

	if actual.BucketName() != name {
		t.Fatalf("Event should have bucket name '%v'. Was '%v'.", name, actual.Namespace())
	}

	if actual.Dynamic() != dyn {
		t.Fatalf("Event should have dynamic='%v'. Was '%v'.", dyn, actual.Dynamic())
	}

	if actual.EventType() != eventType {
		t.Fatalf("Event should have type '%v'. Was '%v'.", eventType, actual.EventType())
	}

	if actual.NumTokens() != tokens {
		t.Fatalf("Event should have tokens '%v'. Was '%v'.", tokens, actual.NumTokens())
	}

	if actual.WaitTime() != waitTime {
		t.Fatalf("Event should have wait time '%v'. Was '%v'.", waitTime, actual.WaitTime())
	}
}

func clearEvents(numEvents int) {
	eventsLeft := numEvents
	for _ = range events {
		eventsLeft--
		if eventsLeft == 0 {
			return
		}
	}
}
