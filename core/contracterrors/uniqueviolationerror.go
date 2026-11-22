package contracterrors

import (
	"errors"

	"github.com/vnafikov/common/validation"
)

type UniqueViolationError struct {
	Validator validation.Validator
	Err       error
}

func (e *UniqueViolationError) Error() string {
	return formatError(ErrUniqueViolation, e.Err)
}

func (e *UniqueViolationError) Unwrap() error {
	return e.Err
}

func (*UniqueViolationError) Is(target error) bool {
	return errors.Is(ErrUniqueViolation, target)
}
