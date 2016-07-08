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
	"github.com/maniksurtani/quotaservice/rpc/grpc"
)

func main() {
	cfg := config.NewDefaultServiceConfig()
	ns := config.NewDefaultNamespaceConfig("test.namespace")
	ns.DynamicBucketTemplate = config.NewDefaultBucketConfig("")
	ns.DynamicBucketTemplate.Size = 100000000000
	ns.DynamicBucketTemplate.FillRate = 100000000
	b := config.NewDefaultBucketConfig("xyz")
	config.AddBucket(ns, b)
	config.AddNamespace(cfg, ns)

	ns = config.NewDefaultNamespaceConfig("test.namespace2")
	ns.DefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	b = config.NewDefaultBucketConfig("xyz")
	config.AddBucket(ns, b)
	config.AddNamespace(cfg, ns)

	server := quotaservice.New(memory.NewBucketFactory(),
		config.NewMemoryConfigPersister(),
		cfg,
		grpc.New("localhost:10990"))
	server.Start()

	// Serve Admin Console
	sm := http.NewServeMux()
	server.ServeAdminConsole(sm, "admin/public", true)
	go func() { http.ListenAndServe("localhost:8080", sm) }()

	// Block until SIGTERM, SIGKILL or SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	var shutdown sync.WaitGroup
	shutdown.Add(1)

	go func() {
		<-sigs
		shutdown.Done()
	}()

	shutdown.Wait()
	server.Stop()
}
