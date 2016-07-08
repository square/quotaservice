// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"net/http"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/logging"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// The Server interface is what you get when you create a new quotaservice.
type Server interface {
	Start() (bool, error)
	Stop() (bool, error)
	SetLogger(logger logging.Logger)
	ServeAdminConsole(*http.ServeMux, string, bool)
	SetListener(listener Listener, eventQueueBufSize int)
}

func NewWithDefaultConfig(bucketFactory BucketFactory, rpcEndpoints ...RpcEndpoint) Server {
	return New(bucketFactory, config.NewMemoryConfigPersister(), &pb.ServiceConfig{}, rpcEndpoints...)
}

// New creates a new quotaservice server.
func New(bucketFactory BucketFactory, persister config.ConfigPersister, configs *pb.ServiceConfig, rpcEndpoints ...RpcEndpoint) Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}

	s := &server{
		persister:     persister,
		bucketFactory: bucketFactory,
		rpcEndpoints:  rpcEndpoints}
	s.createBucketContainer(configs)
	return s
}
