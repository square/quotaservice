// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/test"
)

type dummyEndpoint struct {}
func (d *dummyEndpoint) Init(qs QuotaService) {}
func (d *dummyEndpoint) Start() {}
func (d *dummyEndpoint) Stop() {}


func TestWithNoRpcs(t *testing.T) {
	test.ExpectingPanic(t, func() {
		New(configs.NewDefaultServiceConfig(), memory.NewBucketFactory())
	})
}

func TestValidServer(t *testing.T) {
	s := New(configs.NewDefaultServiceConfig(), memory.NewBucketFactory(), &dummyEndpoint{})
	s.Start()
	defer s.Stop()

	if s.Metrics() == nil {
		t.Fatal("Expected a Metrics instance")
	}
}
