// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package main

import (
	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/rpc/grpc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := configs.NewDefaultServiceConfig()
	cfg.Namespaces["test.namespace"] = configs.NewDefaultNamespaceConfig()
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate.Size = 100000000000
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate.FillRate = 100000000
	cfg.Namespaces["test.namespace"].Buckets["xyz"] = configs.NewDefaultBucketConfig()
	cfg.Namespaces["test.namespace2"] = configs.NewDefaultNamespaceConfig()
	cfg.Namespaces["test.namespace2"].DefaultBucket = configs.NewDefaultBucketConfig()
	cfg.Namespaces["test.namespace2"].Buckets["xyz"] = configs.NewDefaultBucketConfig()

	server := quotaservice.New(cfg, memory.NewBucketFactory(), grpc.New("localhost:10990"))
	server.Start()

	// Serve Admin Console
	sm := http.NewServeMux()
	server.ServeAdminConsole(sm)
	http.ListenAndServe("localhost:8080", sm)

	// Block until SIGTERM, SIGKILL or SIGINT
	sigs := make(chan os.Signal, 1)
	shutdown := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	go func() {
		<-sigs
		shutdown <- true
	}()
	<-shutdown
	server.Stop()
}
