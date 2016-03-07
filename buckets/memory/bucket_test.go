// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package memory

import (
	"os"
	"testing"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets"
)

var factory = NewBucketFactory()

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	os.Exit(r)
}

func setUp() {
	factory.Init(quotaservice.NewDefaultServiceConfig())
}

func TestTokenAcquisition(t *testing.T) {
	bucket := factory.NewBucket("memory", "memory", quotaservice.NewDefaultBucketConfig(), false)
	buckets.TestTokenAcquisition(t, bucket)
}

func TestGC(t *testing.T) {
	buckets.TestGC(t, factory, "memory")
}
