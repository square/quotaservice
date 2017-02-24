// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/events"
	"github.com/maniksurtani/quotaservice/logging"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

// bucketContainer is a holder for configurations and bucket factories.
type bucketContainer struct {
	cfg           *pbconfig.ServiceConfig
	bf            BucketFactory
	n             notifier
	namespaces    map[string]*namespace
	defaultBucket Bucket
	r             *reaper
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
	Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration, success bool)
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

func (ns *namespace) exists(name string) bool {
	ns.RLock()
	defer ns.RUnlock()

	return ns.buckets[name] != nil
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
func NewBucketContainer(cfg *pbconfig.ServiceConfig, bf BucketFactory, n notifier, r config.ReaperConfig) (bc *bucketContainer) {
	bc = &bucketContainer{
		cfg:        cfg,
		bf:         bf,
		n:          n,
		namespaces: make(map[string]*namespace)}

	bc.r = newReaper(bc, r)

	if cfg.GlobalDefaultBucket != nil {
		bc.createGlobalDefaultBucket(cfg.GlobalDefaultBucket)
	}

	for name, nsCfg := range cfg.Namespaces {
		if nsCfg.Name == "" {
			nsCfg.Name = name
		}

		bc.createNamespace(nsCfg)
	}

	return
}

func (bc *bucketContainer) createNamespace(nsCfg *pbconfig.NamespaceConfig) error {
	if _, exists := bc.namespaces[nsCfg.Name]; exists {
		return errors.New("Namespace " + nsCfg.Name + " already exists.")
	}

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

	return nil
}

func (bc *bucketContainer) createGlobalDefaultBucket(cfg *pbconfig.BucketConfig) error {
	if bc.defaultBucket != nil {
		return errors.New("Global default bucket already exists")
	}

	bc.defaultBucket = bc.bf.NewBucket(config.GlobalNamespace, config.DefaultBucketName, cfg, false)
	return nil
}

// FindBucket locates a bucket for a given name and namespace. If the namespace doesn't exist, and
// if a global default bucket is configured, it will be used. If the namespace is available but the
// named bucket doesn't exist, it will either use a namespace-scoped default bucket if available, or
// a dynamic bucket is created if enabled (and space for more dynamic buckets is available). If all
// fails, this function returns nil. This function is thread-safe, and may lazily create dynamic
// buckets or re-create statically defined buckets that have been invalidated.
func (bc *bucketContainer) FindBucket(namespace string, bucketName string) (Bucket, error) {
	ns := bc.namespaces[namespace]
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
			logging.Printf("Bucket %v:%v numDynamicBuckets=%v maxDynamicBuckets=%v. Not creating more dynamic buckets.",
				namespace, bucketName, ns.dynamicBucketCount, ns.cfg.MaxDynamicBuckets)
			return nil
		}

		dyn = true
		bCfg = ns.cfg.DynamicBucketTemplate
	}

	return bc.createNewNamedBucketFromCfg(namespace, bucketName, ns, bCfg, dyn)
}

func (bc *bucketContainer) countDynamicBuckets(namespace string) int32 {
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

	bucket, _ = bc.r.applyWatch(bucket, namespace, bucketName, bCfg)
	ns.buckets[bucketName] = bucket

	if dyn {
		ns.dynamicBucketCount++
	}

	bucket.ReportActivity()
	return bucket
}

func (bc *bucketContainer) NamespaceExists(namespace string) bool {
	_, exists := bc.namespaces[namespace]
	return exists
}

func (bc *bucketContainer) Exists(namespace, name string) bool {
	if ns, exists := bc.namespaces[namespace]; exists {
		ns.RLock()
		defer ns.RUnlock()
		_, bucketExists := ns.buckets[name]
		return bucketExists
	}

	return false
}

func (bc *bucketContainer) removeBucket(namespace, bucket string) bool {
	ns := bc.namespaces[namespace]
	if ns != nil {
		ns.removeBucket(bucket)
		return true
	}
	return false
}

func (bc *bucketContainer) String() string {
	var buffer bytes.Buffer
	if bc.defaultBucket != nil {
		buffer.WriteString("Global default present\n\n")
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
		buffer.WriteString(fmt.Sprintf(" * Namespace: %v\n", nsName))
		if ns.defaultBucket != nil {
			buffer.WriteString("   + Default present\n")
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
			buffer.WriteString(fmt.Sprintf("   + %v\n", bName))
		}
		buffer.WriteString("\n")
	}

	return buffer.String()
}

func (bc *bucketContainer) Stop() {
	bc.r.stop()
}
