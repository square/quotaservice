// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import "time"

// QuotaService is the interface used by RPC subsystems when fielding remote requests for quotas.
type QuotaService interface {
	// Allow will tell you whether the tokens requested in a given namespace and name are available.
	// It will reserve the tokens, and tell the caller how long it would have to wait before the
	// tokens are assumed to be available. In that case, the tokens are reserved, and cannot be put
	// back. Wait times will need to be below the maximum allowed wait time for that namespace and
	// name, and this can be overridden by maxWaitMillisOverride, as long as it maxWaitTimeOverride
	// is set. A returned waitTime of 0 means tokens can be used immediately. Errors indicate
	// tokens could not be obtained, and will contain more context once cast to
	// quotaservice.QoutaServiceError.
	Allow(namespace, name string, tokensRequested int64, maxWaitMillisOverride int64, maxWaitTimeOverride bool) (waitTime time.Duration, err error)
}

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
