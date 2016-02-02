package service
import "errors"

type EmptyBucketPolicyOverride int

const (
	SERVER_DEFAULTS EmptyBucketPolicyOverride = iota
	WAIT
	REJECT
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

func NewError(msg string, reason ErrorReason) QuotaServiceError {
	return QuotaServiceError{error: errors.New(msg), Reason: reason}
}
