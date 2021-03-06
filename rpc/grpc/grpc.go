// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package grpc

import (
	"fmt"
	"net"
	"strings"

	"time"

	"github.com/square/quotaservice"
	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/lifecycle"
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type GrpcEndpoint struct {
	hostport      string
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            quotaservice.QuotaService
	producer      events.EventProducer
}

// New creates a new GrpcEndpoint, listening on hostport. Hostport is a string in the form
// "host:port"
func New(hostport string, producer *events.EventProducer) *GrpcEndpoint {
	if producer == nil {
		panic("producer was nil")
	}

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
	}

	grpclog.SetLogger(logging.CurrentLogger())
	g.grpcServer = grpc.NewServer()
	// Each service should be registered
	pb.RegisterQuotaServiceServer(g.grpcServer, g)
	go func() {
		if e := g.grpcServer.Serve(lis); e != nil {
			logging.Fatalf("Cannot start gRPC server. Error %v", e)
		}
	}()
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
		rsp.Status = pb.AllowResponse_REJECTED_INVALID_REQUEST
		return rsp, nil
	}

	var tokensRequested int64 = 1
	if req.TokensRequested > 0 {
		tokensRequested = req.TokensRequested
	}

	wait, dynamic, err := g.qs.Allow(ctx, req.Namespace, req.BucketName, tokensRequested, req.MaxWaitMillisOverride, req.MaxWaitTimeOverride)

	if err != nil {
		if qsErr, ok := err.(quotaservice.QuotaServiceError); ok {
			rsp.Status = toPBStatus(qsErr)
		} else {
			logging.Printf("Caught error %v", err)
			rsp.Status = pb.AllowResponse_REJECTED_SERVER_ERROR
		}

		// If there's a server error, fail open. Otherwise, return the status as is
		if rsp.Status != pb.AllowResponse_REJECTED_SERVER_ERROR {
			return rsp, nil
		}

		g.producer.Emit(events.NewServerErrorEvent(req.Namespace, req.BucketName, dynamic))
	}

	rsp.Status = pb.AllowResponse_OK
	rsp.TokensGranted = req.TokensRequested
	rsp.WaitMillis = wait.Nanoseconds() / int64(time.Millisecond)

	return rsp, nil
}

func invalid(req *pb.AllowRequest) bool {
	return req.BucketName == "" || req.Namespace == ""
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
