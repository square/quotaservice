// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/square/quotaservice/admin"
	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/lifecycle"
	"github.com/square/quotaservice/logging"
	"github.com/square/quotaservice/stats"

	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/square/quotaservice/config"
	pb "github.com/square/quotaservice/protos/config"
)

// Implements the quotaservice.Server interface
type server struct {
	currentStatus     lifecycle.Status
	bucketContainer   *bucketContainer
	bucketFactory     BucketFactory
	rpcEndpoints      []RpcEndpoint
	listener          events.Listener
	statsListener     stats.Listener
	eventQueueBufSize int
	maxJitterMillis   int
	producer          *events.EventProducer
	cfgs              *pb.ServiceConfig
	persister         config.ConfigPersister
	reaperConfig      config.ReaperConfig
	sync.RWMutex      // Embedded mutex
}

func (s *server) String() string {
	return fmt.Sprintf("Quota Server running with status %v", s.currentStatus)
}

func (s *server) Start() (bool, error) {
	bufSize := s.eventQueueBufSize

	if bufSize < 1 {
		bufSize = 1
	}

	// Set up listeners
	s.producer = events.RegisterListener(func(e events.Event) {
		if s.listener != nil {
			s.listener(e)
		}

		if s.statsListener != nil {
			s.statsListener.HandleEvent(e)
		}
	}, bufSize)

	s.createBucketContainer()
	<-s.persister.ConfigChangedWatcher()
	s.readUpdatedConfig(0)
	go s.configListener(s.persister.ConfigChangedWatcher())

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

	// Referencing s.bucketContainer should be guarded
	s.RLock()
	defer s.RUnlock()
	s.bucketContainer.Stop()
	return true, nil
}

func (s *server) Allow(namespace, name string, tokensRequested int64, maxWaitMillisOverride int64, maxWaitTimeOverride bool) (time.Duration, bool, error) {
	s.RLock()
	b, e := s.bucketContainer.FindBucket(namespace, name)
	s.RUnlock()

	if e != nil {
		// Attempted to create a dynamic bucket and failed.
		s.Emit(events.NewBucketMissedEvent(namespace, name, true))
		return 0, true, newError("Cannot create dynamic bucket "+config.FullyQualifiedName(namespace, name), ER_TOO_MANY_BUCKETS)
	}

	if b == nil {
		s.Emit(events.NewBucketMissedEvent(namespace, name, false))
		return 0, false, newError("No such bucket "+config.FullyQualifiedName(namespace, name), ER_NO_BUCKET)
	}

	if b.Config().MaxTokensPerRequest < tokensRequested && b.Config().MaxTokensPerRequest > 0 {
		s.Emit(events.NewTooManyTokensRequestedEvent(namespace, name, b.Dynamic(), tokensRequested))
		return 0, b.Dynamic(), newError(fmt.Sprintf("Too many tokens requested. Bucket %v:%v, tokensRequested=%v, maxTokensPerRequest=%v",
			namespace, name, tokensRequested, b.Config().MaxTokensPerRequest),
			ER_TOO_MANY_TOKENS_REQUESTED)
	}

	maxWaitTime := time.Millisecond
	if maxWaitTimeOverride && maxWaitMillisOverride < b.Config().WaitTimeoutMillis {
		// Use the max wait time override from the request.
		maxWaitTime *= time.Duration(maxWaitMillisOverride)
	} else {
		// Fall back to the max wait time configured on the bucket.
		maxWaitTime *= time.Duration(b.Config().WaitTimeoutMillis)
	}

	w, success := b.Take(tokensRequested, maxWaitTime)

	if !success {
		// Could not claim tokens within the given max wait time
		s.Emit(events.NewTimedOutEvent(namespace, name, b.Dynamic(), tokensRequested))
		return 0, b.Dynamic(), newError(fmt.Sprintf("Timed out waiting on %v:%v", namespace, name), ER_TIMEOUT)
	}

	// The only positive result
	s.Emit(events.NewTokensServedEvent(namespace, name, b.Dynamic(), tokensRequested, w))
	return w, b.Dynamic(), nil
}

func (s *server) ServeAdminConsole(mux *http.ServeMux, assetsDir string, development bool) {
	admin.ServeAdminConsole(s, mux, assetsDir, development)
}

func (s *server) SetLogger(logger logging.Logger) {
	if s.currentStatus == lifecycle.Started {
		panic("Cannot set logger after server has started!")
	}
	logging.SetLogger(logger)
}

func (s *server) SetStatsListener(listener stats.Listener) {
	if s.currentStatus == lifecycle.Started {
		panic("Cannot add listener after server has started!")
	}

	s.statsListener = listener
}

func (s *server) SetListener(listener events.Listener, eventQueueBufSize int) {
	if s.currentStatus == lifecycle.Started {
		panic("Cannot add listener after server has started!")
	}

	if eventQueueBufSize < 1 {
		panic("Event queue buffer size must be greater than 0")
	}

	s.listener = listener
	s.eventQueueBufSize = eventQueueBufSize
}

func (s *server) Emit(e events.Event) {
	if s.producer != nil {
		s.producer.Emit(e)
	}
}

func (s *server) configListener(ch <-chan struct{}) {
	for range ch {
		jitter := 0
		if s.maxJitterMillis != 0 {
			// Pick a random number between 0 and maxJitterMillis
			jitter = rand.Intn(s.maxJitterMillis)
		}
		s.readUpdatedConfig(time.Duration(jitter) * time.Millisecond)
	}
}

func (s *server) readUpdatedConfig(jitter time.Duration) {
	configReader, err := s.persister.ReadPersistedConfig()

	if err != nil {
		logging.Println("error reading persisted config", err)
		return
	}

	newConfig, err := config.Unmarshal(configReader)

	if err != nil {
		logging.Println("error reading marshalled config", err)
		return
	}

	if jitter != 0 {
		time.Sleep(jitter)
	}

	s.updateBucketContainer(newConfig)
}

func (s *server) createBucketContainer() {
	s.Lock()
	defer s.Unlock()

	if s.bucketContainer != nil {
		logging.Fatalf("A bucketcontainer already exists; this shouldn't happen. BucketContainer=%v", s.bucketContainer)
	}
	s.bucketContainer = NewBucketContainer(s.bucketFactory, s, s.reaperConfig)
}

func (s *server) updateBucketContainer(newConfig *pb.ServiceConfig) {
	s.Lock()
	defer s.Unlock()
	s.bucketContainer.Lock()
	defer s.bucketContainer.Unlock()

	// Initialize buckets
	s.bucketFactory.Init(newConfig)

	// If there is no existing config, then this bucket container is brand-new and hasn't been used before.
	firstTime := s.bucketContainer.cfg == nil

	// Set the new config on the the server
	s.cfgs = newConfig

	if firstTime {
		s.bucketContainer.initLocked(newConfig)
		return
	}

	s.bucketContainer.cfg = newConfig
	// Diff existing configs, buckets and namespaces against the new config and see what needs to be evicted

	// Start with the globalDefaultBucket
	var currentDefaultBucketCfg *pb.BucketConfig
	if s.bucketContainer.defaultBucket != nil {
		currentDefaultBucketCfg = s.bucketContainer.defaultBucket.Config()
	}

	if config.DifferentBucketConfigs(currentDefaultBucketCfg, newConfig.GlobalDefaultBucket) {
		if s.bucketContainer.defaultBucket != nil {
			// We need to destroy existing buckets even if we are replacing them.
			s.bucketContainer.defaultBucket.Destroy()
		}

		if newConfig.GlobalDefaultBucket == nil {
			s.bucketContainer.defaultBucket = nil
		} else {
			s.bucketContainer.createGlobalDefaultBucketLocked(s.cfgs.GlobalDefaultBucket)
		}
	}

	// Scan through all namespaces in s.bucketContainer.namespaces and update the config to point to
	// the new instance, *regardless* of whether the config has changed or not. Also, if the config *has*
	// changed, throw away the old namespace and recreate it. We *could* scan all the buckets in a namespace
	// and only recreate the ones that have changed, but this may have little benefit, since the real cost
	// here is with dynamic buckets, and if the namespace config has changed, it's very likely that the
	// change involves the dynamic bucket template.
	for name, ns := range s.bucketContainer.namespaces {
		newNsCfg, exists := newConfig.Namespaces[name]
		if exists {
			if config.DifferentNamespaceConfigs(ns.cfg, newNsCfg) {
				// We need to destroy the old namespace before overwriting.
				ns.destroy()
				// This will overwrite the existing namespace
				s.bucketContainer.createNamespaceLocked(newNsCfg)
			} else {
				// Just correct the config pointer on the old namespace
				ns.swapCfg(newNsCfg)
			}
		} else {
			ns.destroy()
			delete(s.bucketContainer.namespaces, name)
		}
	}

	// Now look for any new namespaces in the new config and add them
	for name, nsCfg := range newConfig.Namespaces {
		if _, exists := s.bucketContainer.namespaces[name]; !exists {
			s.bucketContainer.createNamespaceLocked(nsCfg)
		}
	}
}

func (s *server) updateConfig(user string, updater func(*pb.ServiceConfig) error) error {
	s.Lock()
	clonedCfg := proto.Clone(s.cfgs).(*pb.ServiceConfig)
	currentVersion := clonedCfg.Version
	s.Unlock()

	err := updater(clonedCfg)

	if err != nil {
		return err
	}

	config.ApplyDefaults(clonedCfg)

	clonedCfg.User = user
	clonedCfg.Date = time.Now().Unix()
	clonedCfg.Version = currentVersion + 1

	r, e := config.Marshal(clonedCfg)

	if e != nil {
		return e
	}

	return s.persister.PersistAndNotify(r)
}

// Implements admin.Administrable
func (s *server) Configs() *pb.ServiceConfig {
	s.RLock()
	defer s.RUnlock()
	return s.cfgs
}

func (s *server) UpdateConfig(c *pb.ServiceConfig, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		*clonedCfg = *c
		return nil
	})
}

func (s *server) AddBucket(namespace string, b *pb.BucketConfig, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.CreateBucket(clonedCfg, namespace, b)
	})
}

func (s *server) UpdateBucket(namespace string, b *pb.BucketConfig, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.UpdateBucket(clonedCfg, namespace, b)
	})
}

func (s *server) DeleteBucket(namespace, name, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.DeleteBucket(clonedCfg, namespace, name)
	})
}

func (s *server) AddNamespace(n *pb.NamespaceConfig, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.CreateNamespace(clonedCfg, n)
	})
}

func (s *server) UpdateNamespace(n *pb.NamespaceConfig, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.UpdateNamespace(clonedCfg, n)
	})
}

func (s *server) DeleteNamespace(n, user string) error {
	return s.updateConfig(user, func(clonedCfg *pb.ServiceConfig) error {
		return config.DeleteNamespace(clonedCfg, n)
	})
}

func (s *server) TopDynamicHits(namespace string) []*stats.BucketScore {
	if s.statsListener == nil {
		return nil
	}

	return s.statsListener.TopHits(namespace)
}

func (s *server) TopDynamicMisses(namespace string) []*stats.BucketScore {
	if s.statsListener == nil {
		return nil
	}

	return s.statsListener.TopMisses(namespace)
}

func (s *server) DynamicBucketStats(namespace, bucket string) *stats.BucketScores {
	if s.statsListener == nil {
		return nil
	}

	return s.statsListener.Get(namespace, bucket)
}

func (s *server) HistoricalConfigs() ([]*pb.ServiceConfig, error) {
	configs, err := s.persister.ReadHistoricalConfigs()

	if err != nil {
		return nil, err
	}

	unmarshalledConfigs := make(sortedConfigs, len(configs))

	for i, newConfig := range configs {
		unmarshalledConfig, err := config.Unmarshal(newConfig)

		if err != nil {
			return nil, err
		}

		unmarshalledConfigs[i] = unmarshalledConfig
	}

	sort.Sort(unmarshalledConfigs)

	return unmarshalledConfigs, nil
}

func (s *server) GetServerAdministrable() admin.Administrable {
	return s
}
