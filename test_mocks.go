// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"sync"
	"time"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/events"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

type MockBucket struct {
	sync.RWMutex
	DefaultBucket
	WaitTime              time.Duration
	namespace, bucketName string
	dyn                   bool
	cfg                   *pbconfig.BucketConfig
}

func (b *MockBucket) Take(numTokens int64, maxWaitTime time.Duration) (time.Duration, bool) {
	b.RLock()
	defer b.RUnlock()

	if b.WaitTime > maxWaitTime {
		return 0, false
	}

	return b.WaitTime, true
}
func (b *MockBucket) Config() *pbconfig.BucketConfig {
	return b.cfg
}
func (b *MockBucket) Dynamic() bool {
	return b.dyn
}

type MockBucketFactory struct {
	buckets map[string]*MockBucket
}

func (bf *MockBucketFactory) SetWaitTime(namespace, name string, d time.Duration) {
	bucket := bf.bucket(namespace, name)
	bucket.Lock()
	defer bucket.Unlock()

	bucket.WaitTime = d
}

func (bf *MockBucketFactory) bucket(namespace, name string) *MockBucket {
	fqn := config.FullyQualifiedName(namespace, name)
	bucket := bf.buckets[fqn]
	if bucket == nil {
		panic(fmt.Sprintf("No such bucket %v", fqn))
	}
	return bucket
}

func (bf *MockBucketFactory) Init(cfg *pbconfig.ServiceConfig) {}
func (bf *MockBucketFactory) Client() interface{}              { return nil }
func (bf *MockBucketFactory) NewBucket(namespace, bucketName string, cfg *pbconfig.BucketConfig, dyn bool) Bucket {
	b := &MockBucket{
		WaitTime:   0,
		namespace:  namespace,
		bucketName: bucketName,
		dyn:        dyn,
		cfg:        cfg}

	if bf.buckets == nil {
		bf.buckets = make(map[string]*MockBucket)
	}

	bf.buckets[config.FullyQualifiedName(namespace, bucketName)] = b
	return b
}

type MockEmitter struct {
	Events chan events.Event
}

func (m *MockEmitter) Emit(e events.Event) {
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
func (d *MockEndpoint) Stop()  {}

func NewBucketContainerWithMocks(cfg *pbconfig.ServiceConfig) (*bucketContainer, *MockBucketFactory, *MockEmitter) {
	bf := &MockBucketFactory{}
	e := &MockEmitter{}
	bc := NewBucketContainer(bf, e, NewReaperConfigForTests())
	bc.Init(cfg)

	return bc, bf, e
}

func NewReaperConfigForTests() config.ReaperConfig {
	r := config.NewReaperConfig()
	r.MinFrequency = 100 * time.Millisecond
	r.InitSleep = 100 * time.Millisecond
	return r
}
