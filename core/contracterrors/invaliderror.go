package contracterrors

import (
	"errors"

	"github.com/vnafikov/common/validation"
)

type InvalidError struct {
	Validator validation.Validator
	Err       error
}

func (e *InvalidError) Error() string {
	return formatError(ErrInvalid, e.Err)
}

func (e *InvalidError) Unwrap() error {
	return e.Err
}

func (*InvalidError) Is(target error) bool {
	return errors.Is(ErrInvalid, target)
}
