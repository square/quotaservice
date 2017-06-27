// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package buckets defines interfaces for abstractions of token buckets.
package benchmark

import (
	"fmt"
	"testing"

	"github.com/square/quotaservice"
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/test/helpers"

	pbconfig "github.com/square/quotaservice/protos/config"
)

var benchmarkCfg = func() *pbconfig.ServiceConfig {
	c := config.NewDefaultServiceConfig()
	c.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)

	// Namespace "y"
	ns := config.NewDefaultNamespaceConfig("y")
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig(config.DefaultBucketName)
	ns.MaxDynamicBuckets = 0

	helpers.PanicError(config.AddBucket(ns, config.NewDefaultBucketConfig("y")))
	helpers.PanicError(config.AddNamespace(c, ns))

	return c
}()

var benchmarkContainer, _, _ = quotaservice.NewBucketContainerWithMocks(benchmarkCfg)

func BenchmarkDynamicBucket(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bucket := fmt.Sprintf("new.%d", i)
		_, _ = benchmarkContainer.FindBucket("y", bucket)
	}
}

func BenchmarkFindBucket(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = benchmarkContainer.FindBucket("y", "y")
	}
}
