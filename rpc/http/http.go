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
