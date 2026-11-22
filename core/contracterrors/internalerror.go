package contracterrors

import "errors"

type InternalError struct {
	Err error
}

func (e *InternalError) Error() string {
	return formatError(ErrInternal, e.Err)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

func (*InternalError) Is(target error) bool {
	return errors.Is(ErrInternal, target)
}
