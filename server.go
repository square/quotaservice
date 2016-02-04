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

package quotaservice

import (
	"fmt"
	"time"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice/admin"
	"github.com/maniksurtani/quotaservice/clustering"
)

type Server struct {
	cfgs          *configs.Configs
	currentStatus lifecycle.Status
	stopper       *chan int
	adminServer   *admin.AdminServer
	tokenBuckets  *buckets.TokenBucketsContainer
	bucketFactory buckets.BucketFactory
	rpcEndpoints  []RpcEndpoint
	monitoring    *Monitoring
}

// Constructors
func New(configFile string, bucketFactory buckets.BucketFactory, rpcServers ...RpcEndpoint) *Server {
	return buildServer(configs.ReadConfig(configFile), bucketFactory, rpcServers)
}

func NewWithDefaults(bucketFactory buckets.BucketFactory, rpcServers... RpcEndpoint) *Server {
	return buildServer(configs.NewDefaultConfig(), bucketFactory, rpcServers)
}

func buildServer(config *configs.Configs, bucketFactory buckets.BucketFactory, rpcEndpoints []RpcEndpoint) *Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}
	s := &Server{
		cfgs: config,
		adminServer: admin.NewAdminServer(config.AdminPort),
		bucketFactory: bucketFactory,
		rpcEndpoints: rpcEndpoints}

	// TODO(manik): Metrics? Monitoring? Naming...
	if config.MetricsEnabled {
		s.monitoring = newMonitoring()
	}
	return s
}

func (this *Server) String() string {
	return fmt.Sprintf("Quota Server running on port %v with status %v", this.cfgs.Port, this.currentStatus)
}

func (this *Server) Start() (bool, error) {

	// Initialize buckets
	this.tokenBuckets = buckets.InitBuckets(this.cfgs, this.bucketFactory)
	// Start the admin server
	this.adminServer.Start()

	// Start the RPC servers
	for _, rpcServer := range this.rpcEndpoints {
		rpcServer.Init(this.cfgs, this)
		rpcServer.Start()
	}

	this.currentStatus = lifecycle.Started
	return true, nil
}

func (this *Server) Stop() (bool, error) {
	this.currentStatus = lifecycle.Stopped

	// Stop the admin server
	this.adminServer.Stop()

	// Stop the RPC servers
	for _, rpcServer := range this.rpcEndpoints {
		rpcServer.Stop()
	}

	return true, nil
}

func (this *Server) Allow(bucketName string, tokensRequested int, emptyBucketPolicyOverride EmptyBucketPolicyOverride) (int, error) {
	b := this.tokenBuckets.FindBucket(bucketName)
	if b == nil {
		return 0, newError(fmt.Sprintf("No such bucket named %v", bucketName), ER_NO_SUCH_BUCKET)
	}

	cfg := b.GetConfig()
	granted := false
	nonBlocking := (cfg.RejectIfEmpty && emptyBucketPolicyOverride == EBP_SERVER_DEFAULTS) || emptyBucketPolicyOverride == EBP_REJECT
	if nonBlocking {
		// Non-blocking call
		granted = b.Take(tokensRequested)
	} else {
		// Block until available
		granted = b.TakeBlocking(tokensRequested, time.Duration(cfg.WaitTimeoutMillis) * time.Millisecond)
	}

	if !granted {
		return 0, newError(fmt.Sprintf("Timed out waiting on bucket %v", bucketName), ER_TIMED_OUT_WAITING)
	}

	return tokensRequested, nil
}

func (this *Server) GetMonitoring() *Monitoring {
	return this.monitoring
}

func (this *Server) SetLogger(logger logging.Logger) {
	logging.SetLogger(logger)
}

func (this *Server) SetClustering(clustering clustering.Clustering) {
	// TODO(manik): Implement me
}


