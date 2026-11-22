package render

import (
	"context"

	"github.com/vnafikov/foundation/infrastructure/ws/server"
	"github.com/vnafikov/foundation/json"
)

func SendJSON(ctx context.Context, connection *server.Connection, data any) error {
	payload, err := json.Serialize(ctx, data)
	if err != nil {
		return err
	}
	return connection.Send(ctx, payload)
}
