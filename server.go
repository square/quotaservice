// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"github.com/maniksurtani/quotaservice/admin"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/lifecycle"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice/metrics"
	"net/http"
	"time"
)

type Server interface {
	Start() (bool, error)
	Stop() (bool, error)
	Metrics() metrics.Metrics
	SetLogger(logger logging.Logger)
	ServeAdminConsole(mux *http.ServeMux)
}

type server struct {
	cfgs            *configs.ServiceConfig
	currentStatus   lifecycle.Status
	stopper         *chan int
	bucketContainer *buckets.BucketContainer
	bucketFactory   buckets.BucketFactory
	rpcEndpoints    []RpcEndpoint
	metrics         metrics.Metrics
}

// NewFromFile creates a new quotaservice server.
func New(config *configs.ServiceConfig, bucketFactory buckets.BucketFactory, rpcEndpoints ...RpcEndpoint) Server {
	if len(rpcEndpoints) == 0 {
		panic("Need at least 1 RPC endpoint to run the quota service.")
	}
	s := &server{
		cfgs:          config,
		bucketFactory: bucketFactory,
		rpcEndpoints:  rpcEndpoints}

	if config.MetricsEnabled {
		s.metrics = metrics.New()
	}
	return s
}

func (s *server) String() string {
	return fmt.Sprintf("Quota Server running with status %v", s.currentStatus)
}

func (s *server) Start() (bool, error) {

	// Initialize buckets
	s.bucketFactory.Init(s.cfgs)
	s.bucketContainer = buckets.NewBucketContainer(s.cfgs, s.bucketFactory)

	// Start the RPC servers
	for _, rpcServer := range s.rpcEndpoints {
		rpcServer.Init(s)
		rpcServer.Start()
	}

	if s.cfgs.MetricsEnabled {
		s.metrics = metrics.New()
	}

	s.currentStatus = lifecycle.Started
	return true, nil
}

func (s *server) Stop() (bool, error) {
	s.currentStatus = lifecycle.Stopped

	// Stop the RPC servers
	for _, rpcServer := range s.rpcEndpoints {
		rpcServer.Stop()
	}

	return true, nil
}

func (s *server) Allow(namespace string, name string, tokensRequested int64, maxWaitMillisOverride int64) (granted int64, waitTime time.Duration, err error) {
	b := s.bucketContainer.FindBucket(namespace, name)
	if b == nil {
		err = newError(fmt.Sprintf("No such bucket %v:%v.", namespace, name), ER_NO_SUCH_BUCKET)
		return
	}

	// Timeout
	dur := time.Millisecond
	if maxWaitMillisOverride > -1 && maxWaitMillisOverride < b.Config().WaitTimeoutMillis {
		dur *= time.Duration(maxWaitMillisOverride)
	} else {
		dur *= time.Duration(b.Config().WaitTimeoutMillis)
	}

	waitTime = b.Take(tokensRequested, dur)

	if waitTime < 0 && dur > 0 {
		waitTime = 0
		err = newError(fmt.Sprintf("Timed out waiting on %v:%v", namespace, name), ER_TIMED_OUT_WAITING)
	} else {
		granted = tokensRequested
	}

	return
}

func (s *server) ServeAdminConsole(mux *http.ServeMux) {
	admin.ServeAdminConsole(s, mux)
}

func (s *server) Metrics() metrics.Metrics {
	return s.metrics
}

func (s *server) SetLogger(logger logging.Logger) {
	logging.SetLogger(logger)
}

// Implements admin.Administrable
func (s *server) Configs() *configs.ServiceConfig {
	return s.cfgs
}

func (s *server) BucketContainer() *buckets.BucketContainer {
	return s.bucketContainer
}
