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
)

// BucketContainer is a holder for configurations and bucket factories.
type BucketContainer struct {
	cfg *configs.ServiceConfig
	bf  BucketFactory
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

// BucketFactory creates buckets.
type BucketFactory interface {
	// Init initializes the bucket factory.
	Init(cfg *configs.ServiceConfig)

	// NewBucket creates a new bucket.
	NewBucket(namespace string, bucketName string, cfg *configs.BucketConfig) Bucket
}

// NewBucketContainer creates a new bucket container.
func NewBucketContainer(cfg *configs.ServiceConfig, bf BucketFactory) *BucketContainer {
	return &BucketContainer{cfg: cfg, bf: bf}
}

// FindBucket locates a bucket for a given name and namespace.
func (tb *BucketContainer) FindBucket(namespace string, bucketName string) Bucket {
	// TODO(manik) perform an actual search
	return nil
}
