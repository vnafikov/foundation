package contracterrors

import (
	"errors"

	"github.com/vnafikov/common/validation"
)

type ExclusionViolationError struct {
	Validator validation.Validator
	Err       error
}

func (e *ExclusionViolationError) Error() string {
	return formatError(ErrExclusionViolation, e.Err)
}

func (e *ExclusionViolationError) Unwrap() error {
	return e.Err
}

func (*ExclusionViolationError) Is(target error) bool {
	return errors.Is(ErrExclusionViolation, target)
}
