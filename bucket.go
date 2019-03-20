// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/logging"

	pbconfig "github.com/square/quotaservice/protos/config"
)

// bucketContainer is a holder for configurations and bucket factories.
type bucketContainer struct {
	cfg           *pbconfig.ServiceConfig
	bf            BucketFactory
	n             notifier
	namespaces    map[string]*namespace
	defaultBucket Bucket
	r             *reaper
	sync.RWMutex  // Embedded mutex
}

type namespace struct {
	n                  notifier
	name               string
	cfg                *pbconfig.NamespaceConfig
	buckets            map[string]Bucket
	dynamicBucketCount int32
	defaultBucket      Bucket
	sync.RWMutex       // Embedded mutex
}

type notifier interface {
	Emit(e events.Event)
}

// Bucket is an abstraction of a token bucket.
type Bucket interface {
	// Take retrieves tokens from a token bucket, returning the time, in millis, to wait before
	// the number of tokens becomes available. A return value of 0 would mean no waiting is
	// necessary. Success is true if tokens can be obtained, false if cannot be obtained within
	// the specified maximum wait time.
	Take(ctx context.Context, numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration, success bool)
	Config() *pbconfig.BucketConfig
	// Dynamic indicates whether a bucket is a dynamic one, or one that is statically defined in
	// configuration.
	Dynamic() bool
	// Destroy indicates that a bucket has been removed from the BucketContainer, is no longer
	// reachable, and should clean up any resources it may have open.
	Destroy()
	// ReportActivity indicates that an ActivityChannel is active. This method shouldn't block.
	ReportActivity()
}

type DefaultBucket struct {
}

func (d DefaultBucket) Destroy() {
	// no-op
}

func (d DefaultBucket) ReportActivity() {
	// no-op
}

func (ns *namespace) removeBucket(bucketName string) {
	// Remove this bucket.
	ns.Lock()
	defer ns.Unlock()
	bucket := ns.buckets[bucketName]
	if bucket != nil {
		delete(ns.buckets, bucketName)
		if bucket.Dynamic() {
			ns.dynamicBucketCount--
		}
		ns.n.Emit(events.NewBucketRemovedEvent(ns.name, bucketName, bucket.Dynamic()))
		bucket.Destroy()
	}
}

// destroy calls Destroy() on all buckets in this namespace
func (ns *namespace) destroy() {
	ns.Lock()
	defer ns.Unlock()
	if ns.defaultBucket != nil {
		ns.defaultBucket.Destroy()
	}

	for _, bucket := range ns.buckets {
		bucket.Destroy()
	}
}

// swapCfg swaps the bucket config for the namespace.
func (ns *namespace) swapCfg(newCfg *pbconfig.NamespaceConfig) {
	ns.Lock()
	defer ns.Unlock()
	ns.cfg = newCfg
}

// BucketFactory creates buckets.
type BucketFactory interface {
	// Init initializes the bucket factory.
	Init(cfg *pbconfig.ServiceConfig)

	// NewBucket creates a new bucket.
	NewBucket(namespace, bucketName string, cfg *pbconfig.BucketConfig, dyn bool) Bucket

	// Client is an accessor to the underlying network client, if there is one.
	Client() interface{}
}

// NewBucketContainer creates a new bucket container.
func NewBucketContainer(bf BucketFactory, n notifier, r config.ReaperConfig) (bc *bucketContainer) {
	bc = &bucketContainer{
		bf:         bf,
		n:          n,
		namespaces: make(map[string]*namespace)}

	bc.r = newReaper(bc, r)

	return
}

func (bc *bucketContainer) Init(cfg *pbconfig.ServiceConfig) {
	bc.Lock()
	defer bc.Unlock()

	bc.initLocked(cfg)
}

func (bc *bucketContainer) initLocked(cfg *pbconfig.ServiceConfig) {
	if bc.cfg != nil {
		logging.Fatal("BucketContainer already has a config; cannot be re-initialized")
	}
	bc.cfg = cfg
	if cfg.GlobalDefaultBucket != nil {
		if bc.defaultBucket != nil {
			logging.Fatal("Global default bucket already exists when initializing")
		}
		bc.createGlobalDefaultBucketLocked(cfg.GlobalDefaultBucket)
	}

	for name, nsCfg := range bc.cfg.Namespaces {
		if nsCfg.Name == "" {
			nsCfg.Name = name
		}

		bc.createNamespaceLocked(nsCfg)
	}
}

func (bc *bucketContainer) createNamespaceLocked(nsCfg *pbconfig.NamespaceConfig) {
	nsp := &namespace{n: bc.n, name: nsCfg.Name, cfg: nsCfg, buckets: make(map[string]Bucket)}
	if nsCfg.DefaultBucket != nil {
		nsp.defaultBucket = bc.bf.NewBucket(nsCfg.Name, config.DefaultBucketName, nsCfg.DefaultBucket, false)
	}

	nsp.Lock()
	defer nsp.Unlock()

	for bucketName, bucketCfg := range nsCfg.Buckets {
		bc.createNewNamedBucketFromCfg(nsCfg.Name, bucketName, nsp, bucketCfg, false)
	}

	bc.namespaces[nsCfg.Name] = nsp
}

func (bc *bucketContainer) createGlobalDefaultBucketLocked(cfg *pbconfig.BucketConfig) {
	bc.defaultBucket = bc.bf.NewBucket(config.GlobalNamespace, config.DefaultBucketName, cfg, false)
}

// FindBucket locates a bucket for a given name and namespace. If the namespace doesn't exist, and
// if a global default bucket is configured, it will be used. If the namespace is available but the
// named bucket doesn't exist, it will either use a namespace-scoped default bucket if available, or
// a dynamic bucket is created if enabled (and space for more dynamic buckets is available). If all
// fails, this function returns nil. This function is thread-safe, and may lazily create dynamic
// buckets or re-create statically defined buckets that have been invalidated.
func (bc *bucketContainer) FindBucket(namespace string, bucketName string) (Bucket, error) {
	bc.RLock()
	ns := bc.namespaces[namespace]
	bc.RUnlock()

	var bucket Bucket
	var err error
	reportActivity := true

	if ns == nil {
		// Namespace doesn't exist. Use default bucket if possible.
		bucket = bc.defaultBucket
	} else {
		// Check if the precise bucket exists.
		ns.RLock()
		bucket = ns.buckets[bucketName]
		ns.RUnlock()

		if bucket == nil {
			if ns.cfg.DynamicBucketTemplate != nil {
				// Double-checked locking is safe in Golang, since acquiring locks (read or write)
				// have the same effect as volatile in Java, causing a memory fence being crossed.
				ns.Lock()
				defer ns.Unlock()
				// need to check if an instance has been created concurrently.
				bucket = ns.buckets[bucketName]
				if bucket == nil {
					reportActivity = false // createNewNamedBucket will report activity
					bucket = bc.createNewNamedBucket(namespace, bucketName, ns)
					if bucket == nil {
						err = errors.New("Cannot create dynamic bucket")
					}
				}
			} else {
				// Try a default for the namespace.
				bucket = ns.defaultBucket
			}
		}
	}

	if bucket != nil && reportActivity {
		bucket.ReportActivity()
	}

	return bucket, err
}

// createNewNamedBucket creates a new, named bucket. May return nil if the named bucket is dynamic,
// and the namespace has already reached its maxDynamicBuckets setting.
func (bc *bucketContainer) createNewNamedBucket(namespace, bucketName string, ns *namespace) Bucket {
	bCfg := ns.cfg.Buckets[bucketName]
	dyn := false
	if bCfg == nil {
		// Dynamic.
		if ns.dynamicBucketCount >= ns.cfg.MaxDynamicBuckets && ns.cfg.MaxDynamicBuckets > 0 {
			logging.Infof("Bucket %v:%v numDynamicBuckets=%v maxDynamicBuckets=%v. Not creating more dynamic buckets.",
				namespace, bucketName, ns.dynamicBucketCount, ns.cfg.MaxDynamicBuckets)
			return nil
		}

		dyn = true
		bCfg = ns.cfg.DynamicBucketTemplate
	}

	return bc.createNewNamedBucketFromCfg(namespace, bucketName, ns, bCfg, dyn)
}

func (bc *bucketContainer) countDynamicBuckets(namespace string) int32 {
	bc.RLock()
	defer bc.RUnlock()

	var c int32
	for _, b := range bc.namespaces[namespace].buckets {
		if b.Dynamic() {
			c++
		}
	}
	return c
}

func (bc *bucketContainer) createNewNamedBucketFromCfg(namespace, bucketName string, ns *namespace, bCfg *pbconfig.BucketConfig, dyn bool) Bucket {
	bc.n.Emit(events.NewBucketCreatedEvent(namespace, bucketName, dyn))
	var bucket Bucket
	bucket = bc.bf.NewBucket(namespace, bucketName, bCfg, dyn)

	if bucket == nil {
		// TODO(manik) why would this ever happen? Should we panic?
		return nil
	}

	if dyn {
		// Apply a watcher if a bucket is dynamic. We don't expire
		// static buckets since FindBucket won't create a new bucket
		// for static buckets. Also, removing idle static buckets
		// won't help much since the number of static buckets is
		// small.
		bucket, _ = bc.r.applyWatch(bucket, namespace, bucketName, bCfg)
		ns.dynamicBucketCount++
	}
	ns.buckets[bucketName] = bucket

	bucket.ReportActivity()
	return bucket
}

func (bc *bucketContainer) NamespaceExists(namespace string) bool {
	bc.RLock()
	defer bc.RUnlock()

	_, exists := bc.namespaces[namespace]
	return exists
}

func (bc *bucketContainer) Exists(namespace, name string) bool {
	bc.RLock()
	defer bc.RUnlock()

	if ns, exists := bc.namespaces[namespace]; exists {
		ns.RLock()
		defer ns.RUnlock()
		_, bucketExists := ns.buckets[name]
		return bucketExists
	}

	return false
}

func (bc *bucketContainer) removeBucket(namespace, bucket string) bool {
	bc.RLock()
	ns := bc.namespaces[namespace]
	bc.RUnlock()

	if ns != nil {
		ns.removeBucket(bucket)
		return true
	}
	return false
}

func (bc *bucketContainer) String() string {
	bc.RLock()
	defer bc.RUnlock()

	var buffer bytes.Buffer
	if bc.defaultBucket != nil {
		_, _ = buffer.WriteString("Global default present\n\n")
	}

	sortedNamespaces := make([]string, len(bc.namespaces))
	i := 0
	for nsName := range bc.namespaces {
		sortedNamespaces[i] = nsName
		i++
	}

	sort.Strings(sortedNamespaces)

	for _, nsName := range sortedNamespaces {
		ns := bc.namespaces[nsName]
		_, _ = buffer.WriteString(fmt.Sprintf(" * Namespace: %v\n", nsName))
		if ns.defaultBucket != nil {
			_, _ = buffer.WriteString("   + Default present\n")
		}

		// Sort buckets
		sortedBuckets := make([]string, len(ns.buckets))
		j := 0
		for bName := range ns.buckets {
			sortedBuckets[j] = bName
			j++
		}

		sort.Strings(sortedBuckets)

		for _, bName := range sortedBuckets {
			_, _ = buffer.WriteString(fmt.Sprintf("   + %v\n", bName))
		}
		_, _ = buffer.WriteString("\n")
	}

	return buffer.String()
}

func (bc *bucketContainer) Stop() {
	bc.r.stop()
}
