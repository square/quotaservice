// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
	"github.com/maniksurtani/quotaservice/test/helpers"
)

func TestWithNoRpcs(t *testing.T) {
	helpers.ExpectingPanic(t, func() {
		New(&MockBucketFactory{}, &config.MemoryConfigPersister{}, &pb.ServiceConfig{})
	})
}

func TestValidServer(t *testing.T) {
	s := New(&MockBucketFactory{}, &config.MemoryConfigPersister{}, &pb.ServiceConfig{}, &MockEndpoint{})
	s.Start()
	defer s.Stop()
}
