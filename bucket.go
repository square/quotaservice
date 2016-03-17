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
	"github.com/maniksurtani/quotaservice/logging"
)

// bucketContainer is a holder for configurations and bucket factories.
type bucketContainer struct {
	cfg           *config.ServiceConfig
	bf            BucketFactory
	n             notifier
	namespaces    map[string]*namespace
	defaultBucket Bucket
	sync.RWMutex  // Embedded mutex
}

type namespace struct {
	n             notifier
	namespaceName string
	cfg           *config.NamespaceConfig
	buckets       map[string]Bucket
	defaultBucket Bucket
	sync.RWMutex  // Embedded mutex
}

type notifier interface {
	Emit(e Event)
}

// Bucket is an abstraction of a token bucket.
type Bucket interface {
	ActivityReporter
	// Take retrieves tokens from a token bucket, returning the time, in millis, to wait before
	// the number of tokens becomes available. A return value of 0 would mean no waiting is
	// necessary, and a wait time that is less than 0 would mean that no tokens would be available
	// within the max time limit specified.
	Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration)
	Config() *config.BucketConfig
	// Dynamic indicates whether a bucket is a dynamic one, or one that is statically defined in
	// configuration.
	Dynamic() bool
	// Destroy indicates that a bucket has been removed from the BucketContainer, is no longer
	// reachable, and should clean up any resources it may have open.
	Destroy()
}

type ActivityReporter interface {
	ActivityDetected() bool
	ReportActivity()
}

// ActivityChannel is a channel that should be embedded into all bucket implementations. It should
// be constructed using NewActivityChannel(), and activity should be reported on the bucket instance
// using ActivityChannel.ReportActivity(), to ensure it isn't assumed to be inactive and removed
// after a period of time.
type ActivityChannel chan bool

// NewActivityChannel creates a new activity channel for use by bucket implementations.
func NewActivityChannel() ActivityChannel {
	return ActivityChannel(make(chan bool, 1))
}

// ReportActivity indicates that an ActivityChannel is active. This method doesn't block.
func (m ActivityChannel) ReportActivity() {
	select {
	case m <- true:
	// reported activity
	default:
		// Already reported
	}
}

// ActivityDetected tells you if activity has been detected since the last time this method was
// called.
func (m ActivityChannel) ActivityDetected() bool {
	select {
	case <-m:
		return true
	default:
		return false
	}
}

// watch watches a bucket for activity, deleting the bucket if no activity has been detected after
// a given duration.
func (ns *namespace) watch(bucketName string, bucket Bucket, freq time.Duration) {
	if freq <= 0 {
		return
	}

	t := time.NewTicker(freq)

	// Wait for a tick
	for range t.C {
		// Check for activity since last run
		if !bucket.ActivityDetected() || !ns.exists(bucketName) {
			break
		}
	}

	t.Stop()
	ns.removeBucket(bucketName)
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
		ns.n.Emit(newBucketRemovedEvent(ns.namespaceName, bucketName, bucket.Dynamic()))
		bucket.Destroy()
	}
}

// BucketFactory creates buckets.
type BucketFactory interface {
	// Init initializes the bucket factory.
	Init(cfg *config.ServiceConfig)

	// NewBucket creates a new bucket.
	NewBucket(namespace, bucketName string, cfg *config.BucketConfig, dyn bool) Bucket
}

// NewBucketContainer creates a new bucket container.
func NewBucketContainer(cfg *config.ServiceConfig, bf BucketFactory, n notifier) (bc *bucketContainer) {
	bc = &bucketContainer{cfg: cfg, bf: bf, n: n, namespaces: make(map[string]*namespace)}
	bc.Lock()
	defer bc.Unlock()

	if cfg.GlobalDefaultBucket != nil {
		bc.createGlobalDefaultBucket(cfg.GlobalDefaultBucket)
	}

	for name, nsCfg := range cfg.Namespaces {
		if nsCfg.Name == "" {
			nsCfg.Name = name
		}
		bc.createNamespaceUnderLock(nsCfg)
	}
	return
}

func (bc *bucketContainer) createNamespaceUnderLock(nsCfg *config.NamespaceConfig) error {
	if _, exists := bc.namespaces[nsCfg.Name]; exists {
		return errors.New("Namespace " + nsCfg.Name + " already exists.")
	}

	nsp := &namespace{n: bc.n, namespaceName: nsCfg.Name, cfg: nsCfg, buckets: make(map[string]Bucket)}
	if nsCfg.DefaultBucket != nil {
		nsp.defaultBucket = bc.bf.NewBucket(nsCfg.Name, config.DefaultBucketName, nsCfg.DefaultBucket, false)
	}

	for bucketName, bucketCfg := range nsCfg.Buckets {
		bc.createNewNamedBucketFromCfg(nsCfg.Name, bucketName, nsp, bucketCfg, false)
	}
	bc.namespaces[nsCfg.Name] = nsp

	return nil
}

func (bc *bucketContainer) createGlobalDefaultBucket(cfg *config.BucketConfig) error {
	if bc.defaultBucket != nil {
		return errors.New("Global default bucket already exists")
	}
	bc.defaultBucket = bc.bf.NewBucket(config.GlobalNamespace, config.DefaultBucketName, cfg, false)
	return nil
}

// findBucket locates a bucket for a given name and namespace. If the namespace doesn't exist, and
// if a global default bucket is configured, it will be used. If the namespace is available but the
// named bucket doesn't exist, it will either use a namespace-scoped default bucket if available, or
// a dynamic bucket is created if enabled (and space for more dynamic buckets is available). If all
// fails, this function returns nil. This function is thread-safe, and may lazily create dynamic
// buckets or re-create statically defined buckets that have been invalidated.
func (bc *bucketContainer) FindBucket(namespace string, bucketName string) (bucket Bucket, err error) {
	bc.RLock()
	ns := bc.namespaces[namespace]
	bc.RUnlock()

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
					bucket = bc.createNewNamedBucket(namespace, bucketName, ns)
					if bucket == nil {
						err = errors.New("Cannot create dynamic bucket")
						return
					}
				}
			} else {
				// Try a default for the namespace.
				bucket = ns.defaultBucket
			}
		}
	}

	if bucket != nil {
		bucket.ReportActivity()
	}

	return
}

// createNewNamedBucket creates a new, named bucket. May return nil if the named bucket is dynamic,
// and the namespace has already reached its maxDynamicBuckets setting.
func (bc *bucketContainer) createNewNamedBucket(namespace, bucketName string, ns *namespace) Bucket {
	bCfg := ns.cfg.Buckets[bucketName]
	dyn := false
	if bCfg == nil {
		// Dynamic.
		numDynamicBuckets := bc.countDynamicBuckets(namespace)
		if numDynamicBuckets >= ns.cfg.MaxDynamicBuckets && ns.cfg.MaxDynamicBuckets > 0 {
			logging.Printf("Bucket %v:%v numDynamicBuckets=%v maxDynamicBuckets=%v. Not creating more dynamic buckets.",
				namespace, bucketName, numDynamicBuckets, ns.cfg.MaxDynamicBuckets)
			return nil
		}

		dyn = true
		bCfg = ns.cfg.DynamicBucketTemplate
	}

	return bc.createNewNamedBucketFromCfg(namespace, bucketName, ns, bCfg, dyn)
}

func (bc *bucketContainer) countDynamicBuckets(namespace string) int {
	c := 0
	for _, b := range bc.namespaces[namespace].buckets {
		if b.Dynamic() {
			c++
		}
	}
	return c
}

func (bc *bucketContainer) createNewNamedBucketFromCfg(namespace, bucketName string, ns *namespace, bCfg *config.BucketConfig, dyn bool) Bucket {
	bc.n.Emit(newBucketCreatedEvent(namespace, bucketName, dyn))
	bucket := bc.bf.NewBucket(namespace, bucketName, bCfg, dyn)
	ns.buckets[bucketName] = bucket
	bucket.ReportActivity()

	if bucketName != config.DefaultBucketName {
		go ns.watch(bucketName, bucket, time.Duration(bCfg.MaxIdleMillis)*time.Millisecond)
	}
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

func (bc *bucketContainer) deleteBucket(namespace, name string) error {
	ns := bc.namespaces[namespace]
	if ns != nil {
		if name == config.DefaultBucketName {
			ns.defaultBucket.Destroy()
			ns.cfg.DefaultBucket = nil
			ns.defaultBucket = nil
		} else if name == config.DynamicBucketTemplateName {
			ns.cfg.DynamicBucketTemplate = nil
		} else {
			delete(ns.cfg.Buckets, name)
			if ns.buckets[name] != nil {
				ns.removeBucket(name)
			}
		}
	} else {
		if namespace == config.GlobalNamespace {
			// Global.
			if name == config.DefaultBucketName {
				if bc.defaultBucket != nil {
					bc.defaultBucket.Destroy()
					bc.defaultBucket = nil
				}
			} else {
				return errors.New("No such bucket " + name + " on global namespace.")
			}
		} else {
			return errors.New("No such namespace " + namespace + ".")
		}
	}

	return nil
}

func (bc *bucketContainer) deleteNamespace(n string) error {
	bc.Lock()
	defer bc.Unlock()

	nsp := bc.namespaces[n]
	if nsp == nil {
		return errors.New("No such namespace " + n)
	}

	delete(bc.namespaces, n)
	bc.deleteBucket(n, config.DefaultBucketName)
	for b, _ := range nsp.buckets {
		bc.deleteBucket(n, b)
	}

	return nil
}

func (bc *bucketContainer) createNamespace(nsCfg *config.NamespaceConfig) error {
	bc.Lock()
	defer bc.Unlock()

	return bc.createNamespaceUnderLock(nsCfg)
}

func (bc *bucketContainer) String() string {
	var buffer bytes.Buffer
	if bc.defaultBucket != nil {
		buffer.WriteString("Global default present\n\n")
	}

	sortedNamespaces := make([]string, len(bc.namespaces))
	i := 0
	for nsName, _ := range bc.namespaces {
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
		for bName, _ := range ns.buckets {
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
