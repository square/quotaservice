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

	"errors"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// Implements the quotaservice.Server interface
type server struct {
	cfgs              *config.ServiceConfig
	currentStatus     lifecycle.Status
	stopper           *chan int
	bucketContainer   *bucketContainer
	bucketFactory     BucketFactory
	rpcEndpoints      []RpcEndpoint
	listener          Listener
	eventQueueBufSize int
	producer          *EventProducer
	p                 config.ConfigPersister
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

func (s *server) Allow(namespace, name string, tokensRequested int64, maxWaitMillisOverride int64) (time.Duration, error) {
	b, e := s.bucketContainer.FindBucket(namespace, name)
	if e != nil {
		// Attempted to create a dynamic bucket and failed.
		s.Emit(newBucketMissedEvent(namespace, name, true))
		return 0, newError("Cannot create dynamic bucket "+config.FullyQualifiedName(namespace, name), ER_TOO_MANY_BUCKETS)
	}

	if b == nil {
		s.Emit(newBucketMissedEvent(namespace, name, false))
		return 0, newError("No such bucket "+config.FullyQualifiedName(namespace, name), ER_NO_BUCKET)
	}

	if b.Config().MaxTokensPerRequest < tokensRequested && b.Config().MaxTokensPerRequest > 0 {
		s.Emit(newTooManyTokensRequestedEvent(namespace, name, b.Dynamic(), tokensRequested))
		return 0, newError(fmt.Sprintf("Too many tokens requested. Bucket %v:%v, tokensRequested=%v, maxTokensPerRequest=%v",
			namespace, name, tokensRequested, b.Config().MaxTokensPerRequest),
			ER_TOO_MANY_TOKENS_REQUESTED)
	}

	maxWaitTime := time.Millisecond
	if maxWaitMillisOverride > -1 && maxWaitMillisOverride < b.Config().WaitTimeoutMillis {
		// Use the max wait time override from the request.
		maxWaitTime *= time.Duration(maxWaitMillisOverride)
	} else {
		// Fall back to the max wait time configured on the bucket.
		maxWaitTime *= time.Duration(b.Config().WaitTimeoutMillis)
	}

	w, success := b.Take(tokensRequested, maxWaitTime)

	if !success {
		// Could not claim tokens within the given max wait time
		s.Emit(newTimedOutEvent(namespace, name, b.Dynamic(), tokensRequested))
		return 0, newError(fmt.Sprintf("Timed out waiting on %v:%v", namespace, name), ER_TIMEOUT)
	}

	// The only positive result
	s.Emit(newTokensServedEvent(namespace, name, b.Dynamic(), tokensRequested, w))
	return w, nil
}

func (s *server) ServeAdminConsole(mux *http.ServeMux, assetsDir string, p config.ConfigPersister) {
	admin.ServeAdminConsole(s, mux, assetsDir)
	s.p = p
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

func (s *server) Emit(e Event) {
	if s.producer != nil {
		s.producer.Emit(e)
	}
}

// Implements admin.Administrable
func (s *server) Configs() *config.ServiceConfig {
	return s.cfgs
}

func (s *server) DeleteBucket(namespace, name string) error {
	err := s.bucketContainer.deleteBucket(namespace, name)
	if err != nil {
		return err
	}

	s.saveUpdatedConfigs()
	return nil
}

func (s *server) AddBucket(namespace string, b *pb.BucketConfig) error {
	if !s.bucketContainer.NamespaceExists(namespace) && namespace != config.GlobalNamespace {
		return errors.New("Namespace doesn't exist")
	}

	if namespace == config.GlobalNamespace {
		err := s.bucketContainer.createGlobalDefaultBucket(config.BucketFromProto(b, nil))
		if err != nil {
			return err
		}
	} else {
		if s.bucketContainer.Exists(namespace, b.Name) {
			return errors.New("Bucket already exists")
		}

		s.bucketContainer.RLock()
		defer s.bucketContainer.RUnlock()
		ns := s.bucketContainer.namespaces[namespace]
		s.bucketContainer.createNewNamedBucketFromCfg(namespace,
			b.Name, ns, config.BucketFromProto(b, ns.cfg), false)
	}

	s.saveUpdatedConfigs()
	return nil
}

func (s *server) UpdateBucket(namespace string, b *pb.BucketConfig) error {
	// Simple delete and add?
	e := s.bucketContainer.deleteBucket(namespace, b.Name)
	if e != nil {
		return e
	}

	return s.AddBucket(namespace, b)
}

func (s *server) DeleteNamespace(n string) error {
	err := s.bucketContainer.deleteNamespace(n)
	if err != nil {
		return err
	}

	s.saveUpdatedConfigs()
	return nil
}

func (s *server) AddNamespace(n *pb.NamespaceConfig) error {
	e := s.bucketContainer.createNamespace(config.NamespaceFromProto(n))
	if e != nil {
		return e
	}
	s.saveUpdatedConfigs()
	return nil
}

func (s *server) UpdateNamespace(n *pb.NamespaceConfig) error {
	err := s.bucketContainer.deleteNamespace(n.Name)
	if err != nil {
		return err
	}

	return s.AddNamespace(n)
}

func (s *server) saveUpdatedConfigs() error {
	if s.p != nil {
		r, e := config.Marshal(s.cfgs)
		if e != nil {
			return e
		}
		return s.p.PersistAndNotify(r)
	}
	return nil
}
