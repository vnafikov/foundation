package response

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vnafikov/foundation/core/contracterrors"
)

func Error(err error) error {
	if _, ok := status.FromError(err); ok {
		return err
	}

	var code codes.Code
	switch {
	case errors.Is(err, contracterrors.ErrUnauthenticated):
		code = codes.Unauthenticated
	case errors.Is(err, contracterrors.ErrForbidden):
		code = codes.PermissionDenied
	case errors.Is(err, contracterrors.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, contracterrors.ErrUniqueViolation):
		code = codes.AlreadyExists
	case errors.Is(err, contracterrors.ErrConflict):
		code = codes.FailedPrecondition
	case errors.Is(err, contracterrors.ErrInvalid):
		code = codes.InvalidArgument
	case errors.Is(err, contracterrors.ErrTooManyRequests):
		code = codes.ResourceExhausted
	case errors.Is(err, context.DeadlineExceeded):
		code = codes.DeadlineExceeded
	case errors.Is(err, context.Canceled):
		code = codes.Canceled
	default:
		code = codes.Internal
	}
	return status.Error(code, "")
}
