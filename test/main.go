// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package main

import (
	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/rpc/grpc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"sync"
)

func main() {
	cfg := quotaservice.NewDefaultServiceConfig()
	ns := quotaservice.NewDefaultNamespaceConfig()
	ns.DynamicBucketTemplate = quotaservice.NewDefaultBucketConfig()
	ns.DynamicBucketTemplate.Size = 100000000000
	ns.DynamicBucketTemplate.FillRate = 100000000
	ns.AddBucket("xyz", quotaservice.NewDefaultBucketConfig())
	cfg.AddNamespace("test.namespace", ns)

	ns = quotaservice.NewDefaultNamespaceConfig()
	ns.DefaultBucket = quotaservice.NewDefaultBucketConfig()
	ns.AddBucket("xyz", quotaservice.NewDefaultBucketConfig())
	cfg.AddNamespace("test.namespace2", ns)

	server := quotaservice.New(cfg, memory.NewBucketFactory(), grpc.New("localhost:10990"))
	server.Start()

	// Serve Admin Console
	sm := http.NewServeMux()
	server.ServeAdminConsole(sm)
	http.ListenAndServe("localhost:8080", sm)

	// Block until SIGTERM, SIGKILL or SIGINT
	sigs := make(chan os.Signal, 1)
	var shutdown sync.WaitGroup
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	go func() {
		<-sigs
		shutdown.Done()
	}()

	shutdown.Wait()
	server.Stop()
}
