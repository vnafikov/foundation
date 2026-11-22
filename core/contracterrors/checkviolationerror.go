package contracterrors

import "errors"

type CheckViolationError struct {
	Err error
}

func (e *CheckViolationError) Error() string {
	return formatError(ErrCheckViolation, e.Err)
}

func (e *CheckViolationError) Unwrap() error {
	return e.Err
}

func (*CheckViolationError) Is(target error) bool {
	return errors.Is(ErrCheckViolation, target)
}
