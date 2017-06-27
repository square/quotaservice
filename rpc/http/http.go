// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package http

import (
	"github.com/square/quotaservice"
	"github.com/square/quotaservice/lifecycle"
)

const defaultPort = 80

// HttpEndpoint is an HTTP-based implementation of an RPC endpoint
type HttpEndpoint struct {
	port          int
	currentStatus lifecycle.Status
	qs            quotaservice.QuotaService
}

func New(port int) *HttpEndpoint {
	return &HttpEndpoint{port: port}
}

func NewDefault() *HttpEndpoint {
	return New(defaultPort)
}

func (h *HttpEndpoint) Init(qs quotaservice.QuotaService) {
	h.qs = qs
}

func (h *HttpEndpoint) Start() {
	h.currentStatus = lifecycle.Started
}

func (h *HttpEndpoint) Stop() {
	h.currentStatus = lifecycle.Stopped
}
