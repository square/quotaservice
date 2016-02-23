/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

// Package buckets defines interfaces for abstractions of token buckets.
package buckets

import (
	"github.com/maniksurtani/quotaservice/configs"
	"time"
	"sync"
	"fmt"
	"bytes"
	"sort"
)

const (
	GLOBAL_NAMESPACE = "___GLOBAL___"
	DEFAULT_BUCKET_NAME = "___DEFAULT_BUCKET___"
)

// BucketContainer is a holder for configurations and bucket factories.
type BucketContainer struct {
	cfg           *configs.ServiceConfig
	bf            BucketFactory
	namespaces    map[string]*namespace
	defaultBucket Bucket
}

// Bucket is an abstraction of a token bucket.
type Bucket interface {
	ActivityReporter
	// Take retrieves tokens from a token bucket, returning the time, in millis, to wait before
	// the number of tokens becomes available. A return value of 0 would mean no waiting is
	// necessary, and a wait time that is less than 0 would mean that no tokens would be available
	// within the max time limit specified.
	Take(numTokens int, maxWaitTime time.Duration) (waitTime time.Duration)
	Config() *configs.BucketConfig
}

type ActivityReporter interface {
	ActivityDetected() bool
	ReportActivity()
}

type ActivityChannel chan bool

func NewActivityChannel() ActivityChannel {
	return ActivityChannel(make(chan bool, 1))
}

func (m ActivityChannel) ReportActivity() {
	select {
	case m <- true:
	// reported activity
	default:
	// Already reported
	}
}

func (m ActivityChannel) ActivityDetected() bool {
	select {
	case <-m:
		return true
	default:
		return false
	}
}

type namespace struct {
	cfg           *configs.NamespaceConfig
	buckets       map[string]Bucket
	defaultBucket Bucket
	sync.RWMutex // Embedded mutex
}

func (ns *namespace) watch(bucketName string, bucket Bucket, freq time.Duration) {
	if freq == 0 {
		return
	}

	t := time.Tick(freq)

	keepRunning := true
	for keepRunning {
		// Wait for a tick
		_ = <-t
		// Check for activity since last run
		keepRunning = bucket.ActivityDetected()
	}

	// Remove this bucket.
	ns.Lock()
	defer ns.Unlock()
	delete(ns.buckets, bucketName)
}

// BucketFactory creates buckets.
type BucketFactory interface {
	// Init initializes the bucket factory.
	Init(cfg *configs.ServiceConfig)

	// NewBucket creates a new bucket.
	NewBucket(namespace, bucketName string, cfg *configs.BucketConfig) Bucket
}

func FullyQualifiedName(namespace, bucketName string) string {
	return fmt.Sprintf("%v:%v", namespace, bucketName)
}

// NewBucketContainer creates a new bucket container.
func NewBucketContainer(cfg *configs.ServiceConfig, bf BucketFactory) (bc *BucketContainer) {
	bc = &BucketContainer{cfg: cfg, bf: bf, namespaces: make(map[string]*namespace)}

	if cfg.GlobalDefaultBucket != nil {
		bc.defaultBucket = bf.NewBucket(GLOBAL_NAMESPACE, DEFAULT_BUCKET_NAME, cfg.GlobalDefaultBucket)
	}

	for nsName, nsCfg := range cfg.Namespaces {
		nsp := &namespace{cfg: nsCfg, buckets: make(map[string]Bucket)}
		if nsCfg.DefaultBucket != nil {
			nsp.defaultBucket = bf.NewBucket(nsName, DEFAULT_BUCKET_NAME, nsCfg.DefaultBucket)
		}

		for bucketName, bucketCfg := range nsCfg.Buckets {
			bc.createNewNamedBucketFromCfg(nsName, bucketName, nsp, bucketCfg)
		}

		bc.namespaces[nsName] = nsp
	}
	return
}

// FindBucket locates a bucket for a given name and namespace.
func (bc *BucketContainer) FindBucket(namespace string, bucketName string) (bucket Bucket) {
	ns := bc.namespaces[namespace]
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
					// TODO(manik) check dynamic bucket count
					bucket = bc.createNewNamedBucket(namespace, bucketName, ns)
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

func (bc *BucketContainer) createNewNamedBucket(namespace, bucketName string, ns *namespace) Bucket {
	bCfg := ns.cfg.Buckets[bucketName]
	if bCfg == nil {
		bCfg = ns.cfg.DynamicBucketTemplate
	}

	return bc.createNewNamedBucketFromCfg(namespace, bucketName, ns, bCfg)

}

func (bc *BucketContainer) createNewNamedBucketFromCfg(namespace, bucketName string, ns *namespace, bCfg *configs.BucketConfig) Bucket {
	bucket := bc.bf.NewBucket(namespace, bucketName, bCfg)
	ns.buckets[bucketName] = bucket
	bucket.ReportActivity()
	go ns.watch(bucketName, bucket, time.Duration(bCfg.MaxIdleMillis) * time.Millisecond)
	return bucket
}

func (bc *BucketContainer) String() string {
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

	for _, nsName := range sortedNamespaces{
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
