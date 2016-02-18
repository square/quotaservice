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
	mutex         sync.RWMutex
}

// Bucket is an abstraction of a token bucket.
type Bucket interface {
	// Take retrieves tokens from a token bucket, returning the time, in millis, to wait before
	// the number of tokens becomes available. A return value of 0 would mean no waiting is
	// necessary, and a wait time that is less than 0 would mean that no tokens would be available
	// within the max time limit specified.
	Take(numTokens int, maxWaitTime time.Duration) (waitTime time.Duration)
	GetConfig() *configs.BucketConfig
}

type namespace struct {
	cfg           *configs.NamespaceConfig
	defaultBucket Bucket
	buckets       map[string]Bucket
}

// BucketFactory creates buckets.
type BucketFactory interface {
	// Init initializes the bucket factory.
	Init(cfg *configs.ServiceConfig)

	// NewBucket creates a new bucket.
	NewBucket(namespace string, bucketName string, cfg *configs.BucketConfig) Bucket
}

// NewBucketContainer creates a new bucket container.
func NewBucketContainer(cfg *configs.ServiceConfig, bf BucketFactory) (bc *BucketContainer) {
	// TODO(manik) start bucket refiller
	// TODO(manik) start bucket reaper
	bc = &BucketContainer{cfg: cfg, bf: bf, namespaces: make(map[string]*namespace)}
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if cfg.GlobalDefaultBucket != nil {
		bc.defaultBucket = bf.NewBucket(GLOBAL_NAMESPACE, DEFAULT_BUCKET_NAME, cfg.GlobalDefaultBucket)
	}

	for nsName, nsCfg := range cfg.Namespaces {
		nsp := &namespace{cfg: nsCfg, buckets: make(map[string]Bucket)}
		if nsCfg.DefaultBucket != nil {
			nsp.defaultBucket = bf.NewBucket(nsName, DEFAULT_BUCKET_NAME, cfg.GlobalDefaultBucket)
		}

		for bucketName, bucketCfg := range nsCfg.Buckets {
			nsp.buckets[bucketName] = bf.NewBucket(nsName, bucketName, bucketCfg)
		}

		bc.namespaces[nsName] = nsp
	}
	return
}

// FindBucket locates a bucket for a given name and namespace.
func (tb *BucketContainer) FindBucket(namespace string, bucketName string) (bucket Bucket) {
	ns := tb.namespaces[namespace]
	if ns == nil {
		// Namespace doesn't exist. Use default bucket if possible.
		bucket = tb.defaultBucket
	} else {

		// Check if the precise bucket exists.
		tb.mutex.RLock()
		bucket = ns.buckets[bucketName]
		tb.mutex.RUnlock()

		if bucket == nil {
			if ns.cfg.DynamicBucketTemplate != nil {
				// Double-checked locking is safe in Golang, since acquiring locks (read or write)
				// have the same effect as volatile in Java, causing a memory fence being crossed.
				tb.mutex.Lock()
				defer tb.mutex.Unlock()
				// need to check if an instance has been created concurrently.
				bucket = ns.buckets[bucketName]
				if bucket == nil {
					// TODO(manik) check dynamic bucket count
					bucket = tb.bf.NewBucket(namespace, bucketName, ns.cfg.DynamicBucketTemplate)
					ns.buckets[bucketName] = bucket
				}
			} else {
				// Try a default for the namespace.
				bucket = ns.defaultBucket
			}
		}
	}

	return
}
