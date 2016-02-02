package quotaservice
import "errors"

type EmptyBucketPolicyOverride int

const (
	EBP_SERVER_DEFAULTS EmptyBucketPolicyOverride = iota
	EBP_WAIT
	EBP_REJECT
)

type ErrorReason int

const (
	ER_NO_SUCH_BUCKET  ErrorReason = iota
	ER_TIMED_OUT_WAITING
	ER_REJECTED
)

type QuotaService interface {
	Allow(bucketName string, tokensRequested int, emptyBucketPolicyOverride EmptyBucketPolicyOverride) (int, error)
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
