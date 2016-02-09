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
import "errors"

type ErrorReason int

const (
	ER_NO_SUCH_BUCKET  ErrorReason = iota
	ER_TIMED_OUT_WAITING
	ER_REJECTED
)

type QuotaService interface {
	Allow(namespace string, name string, tokensRequested int) (granted int, waitTime int64, err error)
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
