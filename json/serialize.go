package json

import (
	"context"
	"encoding/json/v2"
	"log/slog"

	"github.com/vnafikov/foundation/logger"
)

func Serialize(ctx context.Context, data any) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		logger.Error(ctx, err, "json.Marshal", slog.Any("data", data))
	}
	return b, err
}
