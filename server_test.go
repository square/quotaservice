// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"
	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/test/helpers"
)

func TestWithNoRpcs(t *testing.T) {
	helpers.ExpectingPanic(t, func() {
		New(config.NewDefaultServiceConfig(), &MockBucketFactory{})
	})
}

func TestValidServer(t *testing.T) {
	s := New(config.NewDefaultServiceConfig(), &MockBucketFactory{}, &MockEndpoint{})
	s.Start()
	defer s.Stop()
}
