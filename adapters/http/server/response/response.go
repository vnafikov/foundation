package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/response/clienterrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/response/servererrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/response/successful"
)

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, contracterrors.ErrUnauthenticated):
		clienterrors.Unauthorized(w)
	case errors.Is(err, contracterrors.ErrForbidden):
		clienterrors.Forbidden(w)
	case errors.Is(err, contracterrors.ErrNotFound):
		clienterrors.NotFound(w)
	case errors.Is(err, contracterrors.ErrConflict):
		if uniqueViolationError, ok := errors.AsType[*contracterrors.UniqueViolationError](err); ok {
			if len(uniqueViolationError.Validator.Errors) > 0 {
				clienterrors.ConflictWithBody(ctx, w, uniqueViolationError.Validator)
				break
			}

			clienterrors.Conflict(w)
			break
		}

		exclusionViolationError, isExclusionViolation := errors.AsType[*contracterrors.ExclusionViolationError](err)
		if isExclusionViolation && len(exclusionViolationError.Validator.Errors) > 0 {
			clienterrors.ConflictWithBody(ctx, w, exclusionViolationError.Validator)
			break
		}

		clienterrors.Conflict(w)
	case errors.Is(err, contracterrors.ErrInvalid):
		invalidError, isInvalid := errors.AsType[*contracterrors.InvalidError](err)
		if isInvalid && len(invalidError.Validator.Errors) > 0 {
			clienterrors.UnprocessableContentWithBody(ctx, w, invalidError.Validator)
			break
		}

		clienterrors.UnprocessableContent(w)
	case errors.Is(err, contracterrors.ErrTooManyRequests):
		clienterrors.TooManyRequests(w)
	case errors.Is(err, contracterrors.ErrNoUpdateParameters):
		successful.NoContent(w)
	default:
		servererrors.Write(w, err)
	}
}
