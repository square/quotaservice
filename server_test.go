// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"
)

func TestWithNoRpcs(t *testing.T) {
	ExpectingPanic(t, func() {
		New(NewDefaultServiceConfig(), &MockBucketFactory{})
	})
}

func TestValidServer(t *testing.T) {
	s := New(NewDefaultServiceConfig(), &MockBucketFactory{}, &MockEndpoint{})
	s.Start()
	defer s.Stop()
}
