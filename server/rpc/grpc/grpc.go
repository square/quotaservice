package grpc
import (
	"net"
	"fmt"
	"github.com/maniksurtani/qs/quotaservice/protos"
	"github.com/maniksurtani/qs/quotaservice/server/logging"
	"github.com/maniksurtani/qs/quotaservice/server/configs"
	"github.com/maniksurtani/qs/quotaservice/server/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"golang.org/x/net/context"
	"github.com/maniksurtani/qs/quotaservice/server/service"
)

// gRPC-backed implementation of an RPC endpoint
type GrpcEndpoint struct {
	cfgs          *configs.Configs
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            service.QuotaService
}

func (this *GrpcEndpoint) Init(cfgs *configs.Configs, qs service.QuotaService) {
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
	quotaservice.RegisterQuotaServiceServer(this.grpcServer, this)
	go this.grpcServer.Serve(lis)
	this.currentStatus = lifecycle.Started
	logging.Printf("Starting server on port %v", this.cfgs.Port)
	logging.Printf("Server status: %v", this.currentStatus)

}

func (this *GrpcEndpoint) Stop() {
	this.currentStatus = lifecycle.Stopped
}

func (this *GrpcEndpoint) Allow(ctx context.Context, req *quotaservice.AllowRequest) (*quotaservice.AllowResponse, error) {
	rsp := new(quotaservice.AllowResponse)
	// TODO(manik) validate inputs
	granted, err := this.qs.Allow(req.BucketName, int(req.TokensRequested), convert(req.EmptyBucketPolicy))

	if err != nil {
		if qsErr, ok := err.(service.QuotaServiceError); ok {
			switch qsErr.Reason {
			case service.ER_NO_SUCH_BUCKET:
				rsp.Status = quotaservice.AllowResponse_REJECTED
			case service.ER_REJECTED:
				rsp.Status = quotaservice.AllowResponse_REJECTED
			case service.ER_TIMED_OUT_WAITING:
				rsp.Status = quotaservice.AllowResponse_TIMED_OUT
			}
		} else {
			return nil, err
		}
	} else {
		rsp.Status = quotaservice.AllowResponse_OK
		rsp.TokensGranted = int32(granted)
	}
	return rsp, nil
}

func convert(o quotaservice.AllowRequest_EmptyBucketPolicyOverride) service.EmptyBucketPolicyOverride {
	switch o {
	case quotaservice.AllowRequest_REJECT:
		return service.REJECT
	case quotaservice.AllowRequest_WAIT:
		return service.WAIT
	case quotaservice.AllowRequest_SERVER_DEFAULTS:
		return service.SERVER_DEFAULTS
	default:
		panic(fmt.Sprintf("Unknown enum value %+v", o))
	}
}
