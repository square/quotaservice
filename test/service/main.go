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
)

type Clustering interface {
	// Returns true if the current node is the leader
	IsLeader() bool
	// Returns a channel that is used to notify the query service of a membership change.
	MembershipChangeNotificationChannel() chan bool
}

type MetricsHandler interface {
	// Returns true if the current node is the leader
	IsLeader() bool
	// Returns a channel that is used to notify the query service of a membership change.
	MembershipChangeNotificationChannel() chan bool
}


func main() {
	server := quotaservice.New("/tmp/test.yaml", memory.BucketFactory{}, &grpc.GrpcEndpoint{})
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

