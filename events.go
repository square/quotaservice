// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"time"

	"fmt"
	"github.com/maniksurtani/quotaservice/logging"
)

type EventType int

const (
	EVENT_TOKENS_SERVED EventType = iota
	EVENT_TIMEOUT_SERVING_TOKENS
	EVENT_TOO_MANY_TOKENS_REQUESTED
	EVENT_BUCKET_MISS
	EVENT_BUCKET_CREATED
	EVENT_BUCKET_REMOVED
)

var eventNames = []string{
	EVENT_TOKENS_SERVED:             "EVENT_TOKENS_SERVED",
	EVENT_TIMEOUT_SERVING_TOKENS:    "EVENT_TIMEOUT_SERVING_TOKENS",
	EVENT_TOO_MANY_TOKENS_REQUESTED: "EVENT_TOO_MANY_TOKENS_REQUESTED",
	EVENT_BUCKET_MISS:               "EVENT_BUCKET_MISS",
	EVENT_BUCKET_CREATED:            "EVENT_BUCKET_CREATED",
	EVENT_BUCKET_REMOVED:            "EVENT_BUCKET_REMOVED"}

func (et EventType) String() string {
	name := eventNames[et]
	if name == "" {
		panic(fmt.Sprintf("Don't know event %d", et))
	}

	return name
}

type Event interface {
	EventType() EventType
	Namespace() string
	BucketName() string
	Dynamic() bool
	NumTokens() int64
	WaitTime() time.Duration
}

// EventProducer is a hook into the notification system, to inform listeners that certain events
// take place.
type EventProducer struct {
	c chan Event
}

func (e *EventProducer) Emit(event Event) {
	select {
	case e.c <- event:
	// OK
	default:
		logging.Println("Event buffer full; dropping event.")
	}
}

func (ep *EventProducer) notifyListeners(l Listener) {
	for event := range ep.c {
		l(event)
	}
}

type Listener func(details Event)

func registerListener(listener Listener, bufsize int) *EventProducer {
	if listener == nil {
		panic("Cannot register a nil listener")
	}

	ep := &EventProducer{make(chan Event, bufsize)}

	go ep.notifyListeners(listener)

	return ep
}

type namedEvent struct {
	eventType             EventType
	namespace, bucketName string
	dynamic               bool
}

func (n *namedEvent) String() string {
	return fmt.Sprintf("namedEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		n.eventType, n.namespace, n.bucketName, n.dynamic, 0, 0)
}

func (n *namedEvent) EventType() EventType {
	return n.eventType
}

func (n *namedEvent) Namespace() string {
	return n.namespace
}

func (n *namedEvent) BucketName() string {
	return n.bucketName
}

func (n *namedEvent) Dynamic() bool {
	return n.dynamic
}

func (n *namedEvent) NumTokens() int64 {
	return 0
}

func (n *namedEvent) WaitTime() time.Duration {
	return 0
}

type tokenEvent struct {
	*namedEvent
	numTokens int64
}

func (t *tokenEvent) String() string {
	return fmt.Sprintf("tokenEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		t.eventType, t.namespace, t.bucketName, t.dynamic, t.numTokens, 0)
}

func (t *tokenEvent) NumTokens() int64 {
	return t.numTokens
}

type tokenWaitEvent struct {
	*tokenEvent
	waitTime time.Duration
}

func (t *tokenWaitEvent) String() string {
	return fmt.Sprintf("tokenWaitEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		t.eventType, t.namespace, t.bucketName, t.dynamic, t.numTokens, t.waitTime)
}

func (t *tokenWaitEvent) WaitTime() time.Duration {
	return t.waitTime
}

func newTokensServedEvent(namespace, bucketName string, dynamic bool, numTokens int64, waitTime time.Duration) Event {
	return &tokenWaitEvent{
		tokenEvent: &tokenEvent{
			namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TOKENS_SERVED),
			numTokens:  numTokens},
		waitTime: waitTime}
}

func newTimedOutEvent(namespace, bucketName string, dynamic bool, numTokens int64) Event {
	return &tokenEvent{
		namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TIMEOUT_SERVING_TOKENS),
		numTokens:  numTokens}
}

func newTooManyTokensRequestedEvent(namespace, bucketName string, dynamic bool, numTokens int64) Event {
	return &tokenEvent{
		namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TOO_MANY_TOKENS_REQUESTED),
		numTokens:  numTokens}
}

func newBucketMissedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_MISS)
}

func newBucketCreatedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_CREATED)
}

func newBucketRemovedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_REMOVED)
}

func newNamedEvent(namespace, bucketName string, dynamic bool, eventType EventType) *namedEvent {
	return &namedEvent{
		eventType:  eventType,
		namespace:  namespace,
		bucketName: bucketName,
		dynamic:    dynamic}
}
