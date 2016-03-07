// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"testing"
	"time"
	"sync"
)

type MockBucket struct {
	sync.RWMutex
	WaitTime              time.Duration
	Active                bool
	namespace, bucketName string
	dyn                   bool
	cfg                   *BucketConfig
}

func (b *MockBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	b.RLock()
	defer b.RUnlock()

	return b.WaitTime
}
func (b *MockBucket) Config() *BucketConfig {
	return b.cfg
}
func (b *MockBucket) ActivityDetected() bool {
	b.RLock()
	defer b.RUnlock()

	return b.Active
}
func (b *MockBucket) ReportActivity() {}
func (b *MockBucket) Dynamic() bool {
	return b.dyn
}
func (b *MockBucket) Destroy() {}

type MockBucketFactory struct {
	buckets map[string]*MockBucket
}

func (bf *MockBucketFactory) SetWaitTime(namespace, name string, d time.Duration) {
	bucket := bf.bucket(namespace, name)
	bucket.Lock()
	defer bucket.Unlock()

	bucket.WaitTime = d
}

func (bf *MockBucketFactory) SetActive(namespace, name string, active bool) {
	bucket := bf.bucket(namespace, name)
	bucket.Lock()
	defer bucket.Unlock()

	bucket.Active = active
}

func (bf *MockBucketFactory) bucket(namespace, name string) *MockBucket {
	fqn := FullyQualifiedName(namespace, name)
	bucket := bf.buckets[fqn]
	if bucket == nil {
		panic(fmt.Sprintf("No such bucket %v", fqn))
	}
	return bucket
}

func (bf *MockBucketFactory) Init(cfg *ServiceConfig) {}
func (bf *MockBucketFactory) NewBucket(namespace string, bucketName string, cfg *BucketConfig, dyn bool) Bucket {
	b := &MockBucket{sync.RWMutex{}, 0, true, namespace, bucketName, dyn, cfg}
	if bf.buckets == nil {
		bf.buckets = make(map[string]*MockBucket)
	}

	bf.buckets[FullyQualifiedName(namespace, bucketName)] = b
	return b
}

type MockEmitter struct{
	Events chan Event
}

func (m *MockEmitter) Emit(e Event) {
	if m.Events != nil {
		m.Events <- e
	}
}

type MockEndpoint struct {
	QuotaService QuotaService
}

func (d *MockEndpoint) Init(qs QuotaService) {
	d.QuotaService = qs
}
func (d *MockEndpoint) Start() {}
func (d *MockEndpoint) Stop() {}

func ExpectingPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Did not panic()")
		} else {
			fmt.Print(r)
		}
	}()

	f()
}

func NewBucketContainerWithMocks(cfg *ServiceConfig) (*bucketContainer, *MockBucketFactory, *MockEmitter) {
	bf := &MockBucketFactory{}
	e := &MockEmitter{}
	return NewBucketContainer(cfg, bf, e), bf, e
}
