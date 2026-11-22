package protocol

const (
	ErrorCodeUnauthenticated ErrorCode = "unauthenticated"
	ErrorCodeForbidden       ErrorCode = "forbidden"
	ErrorCodeNotFound        ErrorCode = "not_found"
	ErrorCodeInvalid         ErrorCode = "invalid"
	ErrorCodeTooManyRequests ErrorCode = "too_many_requests"
	ErrorCodeInternal        ErrorCode = "internal"
)

type ErrorCode string
