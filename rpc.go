// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

// RpcEndpoint defines a subsystem that listens on a network socket for external systems to
// communicate with the quota service. Endpoints get initialized with a QuotaService interface
// which provides the necessary functionality needed to service requests.
type RpcEndpoint interface {
	// Init will be called before the quotaservice starts, so the RPC subsystem can initialize.
	Init(qs QuotaService)

	// Start will be called after quotaservice has started.
	Start()

	// Stop will be called before the quotaservice stops.
	Stop()
}
