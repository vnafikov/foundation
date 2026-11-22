package contracterrors

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrInvalid         = errors.New("invalid")
	ErrTooManyRequests = errors.New("too many requests")
	ErrInternal        = errors.New("internal error")
	ErrRetryAt         = errors.New("retry at")

	ErrUniqueViolation    = fmt.Errorf("%w: unique violation", ErrConflict)
	ErrCheckViolation     = fmt.Errorf("%w: check violation", ErrConflict)
	ErrExclusionViolation = fmt.Errorf("%w: exclusion violation", ErrConflict)

	ErrNoUpdateParameters = errors.New("no parameters to update")
)
