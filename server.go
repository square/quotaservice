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
)

type Server struct {
	cfgs          *configs.Configs
	currentStatus lifecycle.Status
	stopper       *chan int
	adminServer   *admin.AdminServer
	tokenBuckets  *buckets.TokenBucketsContainer
	rpcEndpoints  []RpcEndpoint
}

// Constructors
func New(configFile string, rpcServers ...RpcEndpoint) *Server {
	return buildServer(configs.ReadConfig(configFile), rpcServers)
}

func NewWithDefaults(rpcServers... RpcEndpoint) *Server {
	return buildServer(configs.NewDefaultConfig(), rpcServers)
}

func buildServer(config *configs.Configs, rpcEndpoints []RpcEndpoint) *Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}
	return &Server{
		cfgs: config,
		adminServer: admin.NewAdminServer(config.AdminPort),
		rpcEndpoints: rpcEndpoints}
}

func (this *Server) String() string {
	return fmt.Sprintf("Quota Server running on port %v with status %v", this.cfgs.Port, this.currentStatus)
}

func (this *Server) Start() (bool, error) {

	// Initialize buckets
	this.tokenBuckets = buckets.InitBuckets(this.cfgs)
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

	waitTime := b.Take(int64(tokensRequested))
	if waitTime > 0 {
		if (b.Cfg.RejectIfEmpty && emptyBucketPolicyOverride == EBP_SERVER_DEFAULTS) || emptyBucketPolicyOverride == EBP_REJECT {
			return 0, newError(fmt.Sprintf("Rejected for bucket %v", bucketName), ER_REJECTED)
		} else if waitTime > (time.Duration(b.Cfg.WaitTimeoutMillis) * time.Millisecond) {
			return 0, newError(fmt.Sprintf("Timed out waiting on bucket %v", bucketName), ER_TIMED_OUT_WAITING)
		} else {
			logging.Printf("Waiting %v", waitTime)
			time.Sleep(waitTime)
		}
	}

	return tokensRequested, nil
}
