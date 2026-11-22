package contracterrors

import "errors"

type UnauthenticatedError struct {
	Err error
}

func (e *UnauthenticatedError) Error() string {
	return formatError(ErrUnauthenticated, e.Err)
}

func formatError(err, wrappedErr error) string {
	return err.Error() + ": " + wrappedErr.Error()
}

func (e *UnauthenticatedError) Unwrap() error {
	return e.Err
}

func (*UnauthenticatedError) Is(target error) bool {
	return errors.Is(ErrUnauthenticated, target)
}
