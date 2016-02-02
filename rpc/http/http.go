package http

import (
	"google.golang.org/grpc"
	"github.com/mohamedattahri/rst"
	"net/http"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice"
)

// HTTP-backed implementation of an RPC endpoint
type HttpEndpoint struct {
	cfgs          *configs.Configs
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            quotaservice.QuotaService
}

type Response struct {
	granted int
}

func (this *HttpEndpoint) Init(cfgs *configs.Configs, qs quotaservice.QuotaService) {
	this.cfgs = cfgs
	this.qs = qs
}

func (this *HttpEndpoint) Start() {
	mux := rst.NewMux()
	mux.Get("/allow/{bucketname:\\s+}/{tokens:\\d+}", func(vars rst.RouteVars, r *http.Request) (rst.Resource, error) {
		name := vars.Get("bucketname")
		tokens := int32(vars.Get("tokens"))
		granted, err := this.qs.Allow(name, tokens, quotaservice.EBP_SERVER_DEFAULTS)
		if err != nil {
			return nil, err
		}

		return &Response{granted: granted}, nil
	})
	this.currentStatus = lifecycle.Started
}

func (this *HttpEndpoint) Stop() {
	this.currentStatus = lifecycle.Stopped
}
