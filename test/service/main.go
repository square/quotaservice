/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package main
import (
	"github.com/maniksurtani/quotaservice"
	"os"
	"os/signal"
	"syscall"
	"github.com/maniksurtani/quotaservice/rpc/grpc"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/configs"
)

func main() {
	cfg := configs.NewDefaultServiceConfig()
	cfg.Namespaces["test.namespace"] = configs.NewDefaultNamespaceConfig()
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate.Size = 100000000000
	cfg.Namespaces["test.namespace"].DynamicBucketTemplate.FillRate = 100000000

	server := quotaservice.New(cfg, memory.BucketFactory{}, grpc.New(10990))
	// server.SetLogging( ... some custom logger ... );
	// server.SetClustering( ... some custom clustering ... )
	_ = server.GetMonitoring()
	server.Start()

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

