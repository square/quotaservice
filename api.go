// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"net/http"

	"github.com/maniksurtani/quotaservice/logging"
)

// The Server interface is what you get when you create a new quotaservice.
type Server interface {
	Start() (bool, error)
	Stop() (bool, error)
	SetLogger(logger logging.Logger)
	ServeAdminConsole(mux *http.ServeMux)
	SetListener(listener Listener)
}

// NewFromFile creates a new quotaservice server.
func New(config *ServiceConfig, bucketFactory BucketFactory, rpcEndpoints ...RpcEndpoint) Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}
	return &server{
		cfgs:          config,
		bucketFactory: bucketFactory,
		rpcEndpoints:  rpcEndpoints}
}
