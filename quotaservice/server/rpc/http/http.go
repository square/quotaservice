package http

import (
	"github.com/maniksurtani/qs/quotaservice/server/configs"
	"github.com/maniksurtani/qs/quotaservice/server/lifecycle"
	"google.golang.org/grpc"
	"github.com/maniksurtani/qs/quotaservice/server/service"
	"github.com/mohamedattahri/rst"
	"github.com/gorilla/mux"
	"net/http"
)

// HTTP-backed implementation of an RPC endpoint
type HttpEndpoint struct {
	cfgs          *configs.Configs
	grpcServer    *grpc.Server
	currentStatus lifecycle.Status
	qs            service.QuotaService
}

type Response struct {
	granted int
}

func (this *HttpEndpoint) Init(cfgs *configs.Configs, qs service.QuotaService) {
	this.cfgs = cfgs
	this.qs = qs
}

func (this *HttpEndpoint) Start() {
	mux := rst.NewMux()
	mux.Get("/allow/{bucketname:\\s+}/{tokens:\\d+}", func(vars rst.RouteVars, r *http.Request) (rst.Resource, error) {
		name := vars.Get("bucketname")
		tokens := int32(vars.Get("tokens"))
		granted, err := this.qs.Allow(name, tokens, service.SERVER_DEFAULTS)
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
