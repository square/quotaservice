// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"errors"
	"time"
)

type ErrorReason int

const (
// Tokens not  available within max wait time
	ER_TIMEOUT ErrorReason = iota

// No valid bucket
	ER_NO_BUCKET

// Dynamic bucket couldn't be created
	ER_TOO_MANY_BUCKETS

// Too many tokens requested
	ER_TOO_MANY_TOKENS_REQUESTED
)

// QuotaService is the interface used by RPC subsystems when fielding remote requests for quotas.
type QuotaService interface {
	// Allow will tell you whether the tokens requested in a given namespace and name are available.
	// It will reserve the tokens, and return the number granted, as well as how long a caller would
	// have to wait before the tokens are assumed to be available. In that case, the tokens are
	// reserved, and cannot be put back. Wait times will need to be below the maximum allowed wait
	// time for that namespace and name, and this can be overridden by maxWaitMillisOverride. Set
	// maxWaitMillisOverride to -1 if you do not wish to override, or 0 if you do not wish to wait
	// at all.
	Allow(namespace string, name string, tokensRequested int64, maxWaitMillisOverride int64) (granted int64, waitTime time.Duration, err error)
}

type QuotaServiceError struct {
	error
	Reason ErrorReason
}

func (e QuotaServiceError) Error() string {
	return e.error.Error()
}

func newError(msg string, reason ErrorReason) QuotaServiceError {
	return QuotaServiceError{error: errors.New(msg), Reason: reason}
}
