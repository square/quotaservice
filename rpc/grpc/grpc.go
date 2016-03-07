// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package grpc

import (
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/logging"
	pb "github.com/maniksurtani/quotaservice/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type GrpcEndpoint struct {
	hostport      string
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            quotaservice.QuotaService
}

// New creates a new GrpcEndpoint, listening on hostport. Hostport is a string in the form
// "host:port"
func New(hostport string) *GrpcEndpoint {
	if !strings.Contains(hostport, ":") {
		panic(fmt.Sprintf("hostport should be in the format 'host:port', but is currently %v",
			hostport))
	}
	return &GrpcEndpoint{hostport: hostport}
}

func (g *GrpcEndpoint) Init(qs quotaservice.QuotaService) {
	g.qs = qs
}

func (g *GrpcEndpoint) Start() {
	lis, err := net.Listen("tcp", g.hostport)
	if err != nil {
		logging.Fatalf("Cannot start server on port %v. Error %v", g.hostport, err)
		panic(fmt.Sprintf("Cannot start server on port %v. Error %v", g.hostport, err))
	}

	grpclog.SetLogger(logging.CurrentLogger())
	g.grpcServer = grpc.NewServer()
	// Each service should be registered
	pb.RegisterQuotaServiceServer(g.grpcServer, g)
	go g.grpcServer.Serve(lis)
	g.currentStatus = lifecycle.Started
	logging.Printf("Starting server on %v", g.hostport)
	logging.Printf("Server status: %v", g.currentStatus)
}

func (g *GrpcEndpoint) Stop() {
	g.currentStatus = lifecycle.Stopped
}

func (g *GrpcEndpoint) Allow(ctx context.Context, req *pb.AllowRequest) (*pb.AllowResponse, error) {
	rsp := new(pb.AllowResponse)
	if invalid(req) {
		logging.Printf("Invalid request %+v", req)
		s := pb.AllowResponse_REJECTED_INVALID_REQUEST
		rsp.Status = &s
		return rsp, nil
	}

	var numTokensRequested int64 = 1
	if req.NumTokensRequested != nil {
		numTokensRequested = req.GetNumTokensRequested()
	}

	var maxWaitMillisOverride int64 = -1
	if req.MaxWaitMillisOverride != nil {
		maxWaitMillisOverride = req.GetMaxWaitMillisOverride()
	}

	granted, wait, err := g.qs.Allow(req.GetNamespace(), req.GetName(), numTokensRequested, maxWaitMillisOverride)
	var status pb.AllowResponse_Status

	if err != nil {
		if qsErr, ok := err.(quotaservice.QuotaServiceError); ok {
			status = toPBStatus(qsErr)
		} else {
			logging.Printf("Caught error %v", err)
			status = pb.AllowResponse_REJECTED_SERVER_ERROR
		}
	} else {
		if wait > 0 {
			status = pb.AllowResponse_OK_WAIT
		} else {
			status = pb.AllowResponse_OK
		}
		rsp.NumTokensGranted = proto.Int64(granted)
		rsp.WaitMillis = proto.Int64(wait.Nanoseconds())
	}
	rsp.Status = &status
	return rsp, nil
}

func invalid(req *pb.AllowRequest) bool {
	// Negative tokens are allowed!
	return req.GetName() == "" || req.GetNamespace() == "" || (req.NumTokensRequested != nil && req.GetNumTokensRequested() == 0)
}

func toPBStatus(qsErr quotaservice.QuotaServiceError) (r pb.AllowResponse_Status) {
	switch qsErr.Reason {
	case quotaservice.ER_NO_BUCKET:
		r = pb.AllowResponse_REJECTED_NO_BUCKET
	case quotaservice.ER_TOO_MANY_BUCKETS:
		r = pb.AllowResponse_REJECTED_TOO_MANY_BUCKETS
	case quotaservice.ER_TOO_MANY_TOKENS_REQUESTED:
		r = pb.AllowResponse_REJECTED_TOO_MANY_TOKENS_REQUESTED
	case quotaservice.ER_TIMEOUT:
		r = pb.AllowResponse_REJECTED_TIMEOUT
	default:
		r = pb.AllowResponse_REJECTED_SERVER_ERROR
	}

	return
}
