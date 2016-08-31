// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"os"
	"testing"
	"time"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/events"
)

var s Server
var qs QuotaService
var eventsChan chan events.Event
var mbf *MockBucketFactory

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	cfg := config.NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)

	// Namespace "dyn"
	ns := config.NewDefaultNamespaceConfig("dyn")
	tpl := config.NewDefaultBucketConfig("")
	tpl.MaxTokensPerRequest = 5
	tpl.MaxIdleMillis = -1
	config.SetDynamicBucketTemplate(ns, tpl)
	ns.MaxDynamicBuckets = 2
	config.AddNamespace(cfg, ns)

	// Namespace "dyn_gc"
	ns = config.NewDefaultNamespaceConfig("dyn_gc")
	tpl = config.NewDefaultBucketConfig("")
	tpl.MaxTokensPerRequest = 5
	tpl.MaxIdleMillis = 100
	config.SetDynamicBucketTemplate(ns, tpl)
	ns.MaxDynamicBuckets = 3
	config.AddNamespace(cfg, ns)

	// Namespace "nodyn"
	ns = config.NewDefaultNamespaceConfig("nodyn")
	b := config.NewDefaultBucketConfig("b")
	b.MaxTokensPerRequest = 10
	config.AddBucket(ns, b)
	config.AddNamespace(cfg, ns)

	mbf = &MockBucketFactory{}
	me := &MockEndpoint{}
	p := config.NewMemoryConfigPersister()
	s = New(mbf, p, cfg, me)
	eventsChan = make(chan events.Event, 100)
	s.SetListener(func(e events.Event) {
		eventsChan <- e
	}, 100)
	s.Start()
	qs = me.QuotaService
}

func TestTokens(t *testing.T) {
	qs.Allow("nodyn", "b", 1, 0, false)
	checkEvent("nodyn", "b", false, events.EVENT_TOKENS_SERVED, 1, 0, <-eventsChan, t)
}

func TestTooManyTokens(t *testing.T) {
	qs.Allow("nodyn", "b", 100, 0, false)
	checkEvent("nodyn", "b", false, events.EVENT_TOO_MANY_TOKENS_REQUESTED, 100, 0, <-eventsChan, t)
}

func TestTimeout(t *testing.T) {
	mbf.SetWaitTime("nodyn", "b", 2*time.Minute)
	qs.Allow("nodyn", "b", 1, 1, false)
	checkEvent("nodyn", "b", false, events.EVENT_TIMEOUT_SERVING_TOKENS, 1, 0, <-eventsChan, t)
	mbf.SetWaitTime("nodyn", "b", 0)
}

func TestWithWait(t *testing.T) {
	mbf.SetWaitTime("nodyn", "b", 2*time.Nanosecond)
	qs.Allow("nodyn", "b", 1, 10, false)
	checkEvent("nodyn", "b", false, events.EVENT_TOKENS_SERVED, 1, 2*time.Nanosecond, <-eventsChan, t)
	mbf.SetWaitTime("nodyn", "b", 0)
}

func TestNoSuchBucket(t *testing.T) {
	qs.Allow("nodyn", "x", 1, 0, false)
	checkEvent("nodyn", "x", false, events.EVENT_BUCKET_MISS, 0, 0, <-eventsChan, t)
}

func TestNewDynBucket(t *testing.T) {
	qs.Allow("dyn", "b", 1, 0, false)
	checkEvent("dyn", "b", true, events.EVENT_BUCKET_CREATED, 0, 0, <-eventsChan, t)
	checkEvent("dyn", "b", true, events.EVENT_TOKENS_SERVED, 1, 0, <-eventsChan, t)
}

func TestTooManyDynBuckets(t *testing.T) {
	n := clearBuckets("dyn")
	qs.Allow("dyn", "c", 1, 0, false)
	qs.Allow("dyn", "d", 1, 0, false)
	clearEvents(4 + n)

	qs.Allow("dyn", "e", 1, 0, false)
	checkEvent("dyn", "e", true, events.EVENT_BUCKET_MISS, 0, 0, <-eventsChan, t)
}

func TestBucketRemoval(t *testing.T) {
	qs.Allow("dyn_gc", "b", 1, 0, false)
	qs.Allow("dyn_gc", "c", 1, 0, false)
	qs.Allow("dyn_gc", "d", 1, 0, false)
	clearEvents(6)

	// GC thread should run every 100ms for this namespace. Make sure it runs at least once.
	time.Sleep(300 * time.Millisecond)

	for i := 0; i < 3; i++ {
		e := <-eventsChan
		checkEvent("dyn_gc", e.BucketName(), true, events.EVENT_BUCKET_REMOVED, 0, 0, e, t)
	}
}

func checkEvent(namespace, name string, dyn bool, eventType events.EventType, tokens int64, waitTime time.Duration, actual events.Event, t *testing.T) {
	if actual == nil {
		t.Fatalf("Expecting event; was nil.")
	}

	if actual.Namespace() != namespace {
		t.Fatalf("Event should have namespace '%v'. Was '%v'. Event %+v.", namespace, actual.Namespace(), actual)
	}

	if actual.BucketName() != name {
		t.Fatalf("Event should have bucket name '%v'. Was '%v'. Event %+v.", name, actual.BucketName(), actual)
	}

	if actual.Dynamic() != dyn {
		t.Fatalf("Event should have dynamic='%v'. Was '%v'. Event %+v.", dyn, actual.Dynamic(), actual)
	}

	if actual.EventType() != eventType {
		t.Fatalf("Event should have type '%v'. Was '%v'. Event %+v.", eventType, actual.EventType(), actual)
	}

	if actual.NumTokens() != tokens {
		t.Fatalf("Event should have tokens '%v'. Was '%v'. Event %+v.", tokens, actual.NumTokens(), actual)
	}

	if actual.WaitTime() != waitTime {
		t.Fatalf("Event should have wait time '%v'. Was '%v'. Event %+v.", waitTime, actual.WaitTime(), actual)
	}
}

func clearEvents(numEvents int) {
	eventsLeft := numEvents
	for _ = range eventsChan {
		eventsLeft--
		if eventsLeft == 0 {
			return
		}
	}
}

func clearBuckets(ns string) int {
	cleared := 0
	namespace := s.(*server).bucketContainer.namespaces[ns]
	for bn, _ := range namespace.buckets {
		namespace.removeBucket(bn)
		cleared++
	}
	return cleared
}
