package contracterrors

import "errors"

type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return formatError(ErrNotFound, e.Err)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func (*NotFoundError) Is(target error) bool {
	return errors.Is(ErrNotFound, target)
}
