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

package quotaservice

import (
	"errors"
	"time"
)

type ErrorReason int

const (
	ER_NO_SUCH_BUCKET ErrorReason = iota
	ER_TIMED_OUT_WAITING
	ER_REJECTED
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
