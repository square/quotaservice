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
// TODO(manik) Implement this package
package http

import (
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice"
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

func (this *HttpEndpoint) Init(qs quotaservice.QuotaService) {
	this.qs = qs
}

func (this *HttpEndpoint) Start() {
	this.currentStatus = lifecycle.Started
}

func (this *HttpEndpoint) Stop() {
	this.currentStatus = lifecycle.Stopped
}
