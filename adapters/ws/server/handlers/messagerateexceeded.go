package handlers

import (
	"context"

	"github.com/vnafikov/foundation/adapters/ws/server/response"
	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/infrastructure/ws/server"
)

// MessageRateExceeded сообщает о превышении лимита входящих сообщений.
func MessageRateExceeded(ctx context.Context, connection *server.Connection) error {
	return response.SendError(ctx, connection, "", contracterrors.ErrTooManyRequests)
}
