// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"
)

type mockEndpoint struct{}

func (d *mockEndpoint) Init(qs QuotaService) {}
func (d *mockEndpoint) Start() {}
func (d *mockEndpoint) Stop() {}

func TestWithNoRpcs(t *testing.T) {
	ExpectingPanic(t, func() {
		New(NewDefaultServiceConfig(), &mockBucketFactory{})
	})
}

func TestValidServer(t *testing.T) {
	s := New(NewDefaultServiceConfig(), &mockBucketFactory{}, &mockEndpoint{})
	s.Start()
	defer s.Stop()
}
