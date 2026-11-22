package contracterrors

import "errors"

type ForbiddenError struct {
	Err error
}

func (e *ForbiddenError) Error() string {
	return formatError(ErrForbidden, e.Err)
}

func (e *ForbiddenError) Unwrap() error {
	return e.Err
}

func (*ForbiddenError) Is(target error) bool {
	return errors.Is(ErrForbidden, target)
}
