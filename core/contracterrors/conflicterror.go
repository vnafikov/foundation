package contracterrors

import "errors"

type ConflictError struct {
	Err error
}

func (e *ConflictError) Error() string {
	return formatError(ErrConflict, e.Err)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}

func (*ConflictError) Is(target error) bool {
	return errors.Is(ErrConflict, target)
}
