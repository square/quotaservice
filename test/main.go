// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"github.com/square/quotaservice/cmd/server/server"
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/test/helpers"
)

func main() {
	cfg := config.NewDefaultServiceConfig()
	ns := config.NewDefaultNamespaceConfig("test.namespace")
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig(config.DynamicBucketTemplateName)
	ns.DynamicBucketTemplate.Size = 100000000000
	ns.DynamicBucketTemplate.FillRate = 100000000
	b := config.NewDefaultBucketConfig("xyz")
	helpers.PanicError(config.AddBucket(ns, b))
	helpers.PanicError(config.AddNamespace(cfg, ns))

	ns = config.NewDefaultNamespaceConfig("test.namespace2")
	ns.DefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	b = config.NewDefaultBucketConfig("xyz")
	helpers.PanicError(config.AddBucket(ns, b))
	helpers.PanicError(config.AddNamespace(cfg, ns))

	server.RunServer(cfg, []string{"--backend_url=memory://"})
}
