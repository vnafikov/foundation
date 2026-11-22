package response

import (
	"context"
	"errors"

	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/infrastructure/ws/server"
	"github.com/vnafikov/foundation/infrastructure/ws/server/presenters"
	"github.com/vnafikov/foundation/infrastructure/ws/server/protocol"
	"github.com/vnafikov/foundation/infrastructure/ws/server/response/render"
)

func SendError(ctx context.Context, connection *server.Connection, channel string, err error) error {
	var code protocol.ErrorCode
	switch {
	case errors.Is(err, contracterrors.ErrUnauthenticated):
		code = protocol.ErrorCodeUnauthenticated
	case errors.Is(err, contracterrors.ErrForbidden):
		code = protocol.ErrorCodeForbidden
	case errors.Is(err, contracterrors.ErrNotFound):
		code = protocol.ErrorCodeNotFound
	case errors.Is(err, contracterrors.ErrInvalid):
		code = protocol.ErrorCodeInvalid
	case errors.Is(err, contracterrors.ErrTooManyRequests):
		code = protocol.ErrorCodeTooManyRequests
	default:
		code = protocol.ErrorCodeInternal
	}
	return render.SendJSON(ctx, connection, presenters.NewErrorResponse(channel, code))
}
