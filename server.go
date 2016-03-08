// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"net/http"
	"time"

	"github.com/maniksurtani/quotaservice/admin"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/logging"
)

// Implements the quotaservice.Server interface
type server struct {
	cfgs              *ServiceConfig
	currentStatus     lifecycle.Status
	stopper           *chan int
	bucketContainer   *bucketContainer
	bucketFactory     BucketFactory
	rpcEndpoints      []RpcEndpoint
	listener          Listener
	eventQueueBufSize int
	producer          *EventProducer
}

func (s *server) String() string {
	return fmt.Sprintf("Quota Server running with status %v", s.currentStatus)
}

func (s *server) Start() (bool, error) {
	// Set up listeners
	if s.listener != nil {
		s.producer = registerListener(s.listener, s.eventQueueBufSize)
	}

	// Initialize buckets
	s.bucketFactory.Init(s.cfgs)
	s.bucketContainer = NewBucketContainer(s.cfgs, s.bucketFactory, s)

	// Start the RPC servers
	for _, rpcServer := range s.rpcEndpoints {
		rpcServer.Init(s)
		rpcServer.Start()
	}

	s.currentStatus = lifecycle.Started
	return true, nil
}

func (s *server) Stop() (bool, error) {
	s.currentStatus = lifecycle.Stopped

	// Stop the RPC servers
	for _, rpcServer := range s.rpcEndpoints {
		rpcServer.Stop()
	}

	return true, nil
}

func (s *server) Allow(namespace string, name string, tokensRequested int64, maxWaitMillisOverride int64) (granted int64, waitTime time.Duration, err error) {
	b, e := s.bucketContainer.FindBucket(namespace, name)
	if e != nil {
		// Attempted to create a dynamic bucket and failed.
		err = newError(fmt.Sprintf("Cannot create dynamic bucket %v:%v.", namespace, name),
			ER_TOO_MANY_BUCKETS)
		s.Emit(newBucketMissedEvent(namespace, name, true))
		return

	}

	if b == nil {
		err = newError(fmt.Sprintf("No such bucket %v:%v.", namespace, name), ER_NO_BUCKET)
		s.Emit(newBucketMissedEvent(namespace, name, false))
		return
	}

	if b.Config().MaxTokensPerRequest < tokensRequested && b.Config().MaxTokensPerRequest > 0 {
		err = newError(fmt.Sprintf("Too many tokens requested. Bucket %v:%v, tokensRequested=%v, maxTokensPerRequest=%v",
			namespace, name, tokensRequested, b.Config().MaxTokensPerRequest),
			ER_TOO_MANY_TOKENS_REQUESTED)
		s.Emit(newTooManyTokensRequestedEvent(namespace, name, b.Dynamic(), tokensRequested))
		return
	}

	// Timeout
	maxWaitTime := time.Millisecond
	if maxWaitMillisOverride > -1 && maxWaitMillisOverride < b.Config().WaitTimeoutMillis {
		maxWaitTime *= time.Duration(maxWaitMillisOverride)
	} else {
		maxWaitTime *= time.Duration(b.Config().WaitTimeoutMillis)
	}

	waitTime = b.Take(tokensRequested, maxWaitTime)

	if waitTime > 0 && maxWaitTime > 0 {
		waitTime = 0
		err = newError(fmt.Sprintf("Timed out waiting on %v:%v", namespace, name), ER_TIMEOUT)
		s.Emit(newTimedOutEvent(namespace, name, b.Dynamic(), tokensRequested))
	} else {
		granted = tokensRequested
		s.Emit(newTokensServedEvent(namespace, name, b.Dynamic(), tokensRequested, waitTime))
	}

	return
}

func (s *server) ServeAdminConsole(mux *http.ServeMux) {
	admin.ServeAdminConsole(s, mux)
}

func (s *server) SetLogger(logger logging.Logger) {
	if s.currentStatus == lifecycle.Started {
		panic("Cannot set logger after server has started!")
	}
	logging.SetLogger(logger)
}

func (s *server) SetListener(listener Listener, eventQueueBufSize int) {
	if s.currentStatus == lifecycle.Started {
		panic("Cannot add listener after server has started!")
	}

	if eventQueueBufSize < 1 {
		panic("Event queue buffer size must be greater than 0")
	}

	s.listener = listener
	s.eventQueueBufSize = eventQueueBufSize
}

// Implements admin.Administrable
func (s *server) Configs() fmt.Stringer {
	return s.cfgs
}

func (s *server) BucketContainer() fmt.Stringer {
	return s.bucketContainer
}

func (s *server) Emit(e Event) {
	if s.producer != nil {
		s.producer.Emit(e)
	}
}
