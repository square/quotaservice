// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"net/http"

	"github.com/square/quotaservice/admin"
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/logging"
	"github.com/square/quotaservice/stats"
)

// The Server interface is what you get when you create a new quotaservice.
type Server interface {
	Start() (bool, error)
	Stop() (bool, error)
	SetLogger(logger logging.Logger)
	ServeAdminConsole(*http.ServeMux, string, bool)
	SetListener(listener events.Listener, eventQueueBufSize int)
	SetStatsListener(listener stats.Listener)
	GetServerAdministrable() admin.Administrable
}

// NewWithDefaultConfig creates a new quotaservice server with an empty in-memory config and default reaper.
func NewWithDefaultConfig(bucketFactory BucketFactory, rpcEndpoints ...RpcEndpoint) Server {
	return New(bucketFactory,
		config.NewMemoryConfig(config.NewDefaultServiceConfig()),
		config.NewReaperConfig(),
		0,
		rpcEndpoints...)
}

// New creates a new quotaservice server.
func New(bucketFactory BucketFactory, persister config.ConfigPersister, reaperConfig config.ReaperConfig, maxCfgReloadJitterMs int, rpcEndpoints ...RpcEndpoint) Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}

	s := &server{
		persister:       persister,
		bucketFactory:   bucketFactory,
		rpcEndpoints:    rpcEndpoints,
		maxJitterMillis: maxCfgReloadJitterMs,
		reaperConfig:    reaperConfig}
	return s
}
