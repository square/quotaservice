// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice/rpc/grpc"
	"github.com/maniksurtani/quotaservice/stats"
	"github.com/maniksurtani/quotaservice/test/helpers"
)

const (
	adminServer = "localhost:8080"
	gRPCServer  = "localhost:10990"
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

	server := quotaservice.New(memory.NewBucketFactory(),
		config.NewMemoryConfig(cfg),
		config.NewReaperConfig(),
		0,
		grpc.New(gRPCServer))
	server.SetStatsListener(stats.NewMemoryStatsListener())
	if _, e := server.Start(); e != nil {
		panic(e)
	}

	// Serve Admin Console
	logging.Printf("Starting admin server on %v\n", adminServer)
	sm := http.NewServeMux()
	server.ServeAdminConsole(sm, "admin/public", true)
	go func() { _ = http.ListenAndServe(adminServer, sm) }()

	// Block until SIGTERM or SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	var shutdown sync.WaitGroup
	shutdown.Add(1)

	go func() {
		<-sigs
		shutdown.Done()
	}()

	shutdown.Wait()
	_, _ = server.Stop()
}
