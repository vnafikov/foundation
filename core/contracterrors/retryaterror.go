package contracterrors

import (
	"errors"
	"time"
)

type RetryAtError struct {
	At  time.Time
	Err error
}

func (e *RetryAtError) Error() string {
	msg := ErrRetryAt.Error() + " " + e.At.Format(time.RFC3339)
	if e.Err == nil {
		return msg
	}
	return msg + ": " + e.Err.Error()
}

func (e *RetryAtError) RetryAt() time.Time {
	return e.At
}

func (e *RetryAtError) Unwrap() error {
	return e.Err
}

func (*RetryAtError) Is(target error) bool {
	return errors.Is(ErrRetryAt, target)
}
