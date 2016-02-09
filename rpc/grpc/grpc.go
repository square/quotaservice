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

package grpc
import (
	"net"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"golang.org/x/net/context"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice"
	qspb "github.com/maniksurtani/quotaservice/protos"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/configs"

	"github.com/golang/protobuf/proto"
)

// gRPC-backed implementation of an RPC endpoint
type GrpcEndpoint struct {
	cfgs          *configs.Configs
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            quotaservice.QuotaService
}

func (this *GrpcEndpoint) Init(cfgs *configs.Configs, qs quotaservice.QuotaService) {
	this.cfgs = cfgs
	this.qs = qs
}

func (this *GrpcEndpoint) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", this.cfgs.Port))
	if err != nil {
		logging.Fatalf("Cannot start server on port %v. Error %v", this.cfgs.Port, err)
		panic(fmt.Sprintf("Cannot start server on port %v. Error %v", this.cfgs.Port, err))
	}

	grpclog.SetLogger(logging.GetLogger())
	this.grpcServer = grpc.NewServer()
	// Each service should be registered
	qspb.RegisterQuotaServiceServer(this.grpcServer, this)
	go this.grpcServer.Serve(lis)
	this.currentStatus = lifecycle.Started
	logging.Printf("Starting server on port %v", this.cfgs.Port)
	logging.Printf("Server status: %v", this.currentStatus)

}

func (this *GrpcEndpoint) Stop() {
	this.currentStatus = lifecycle.Stopped
}

func (this *GrpcEndpoint) Allow(ctx context.Context, req *qspb.AllowRequest) (*qspb.AllowResponse, error) {
	rsp := new(qspb.AllowResponse)
	// TODO(manik) validate inputs
	granted, wait, err := this.qs.Allow(*req.Namespace, *req.Name, int(*req.NumTokensRequested))
	var status qspb.AllowResponse_Status;

	if err != nil {
		if qsErr, ok := err.(quotaservice.QuotaServiceError); ok {
			switch qsErr.Reason {
			case quotaservice.ER_NO_SUCH_BUCKET:
				status = qspb.AllowResponse_REJECTED
			case quotaservice.ER_REJECTED:
				status = qspb.AllowResponse_REJECTED
			case quotaservice.ER_TIMED_OUT_WAITING:
				status = qspb.AllowResponse_REJECTED
			}
		} else {
			return nil, err
		}
	} else {
		status = qspb.AllowResponse_OK
		rsp.NumTokensGranted = proto.Int(granted)
		rsp.WaitMillis = proto.Int64(wait)
	}
	rsp.Status = &status
	return rsp, nil
}
