// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package quotaservice contains interfaces for extension authors.
// E.g., when providing different RPC endpoints to the quota service.
package quotaservice

import (
	"errors"
)

// ErrorReason provides details on why calls to Allow may fail.
type ErrorReason int

const (
	// Tokens not available within max wait time
	ER_TIMEOUT ErrorReason = iota

	// No valid bucket
	ER_NO_BUCKET

	// Dynamic bucket couldn't be created
	ER_TOO_MANY_BUCKETS

	// Too many tokens requested
	ER_TOO_MANY_TOKENS_REQUESTED
)

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
