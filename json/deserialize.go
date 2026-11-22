package json

import (
	"context"
	"encoding/json/v2"
	"log/slog"

	"github.com/vnafikov/foundation/logger"
)

func Deserialize(ctx context.Context, data []byte, out any) error {
	err := json.Unmarshal(data, out)
	if err != nil {
		logger.Error(ctx, err, "json.Unmarshal", slog.String("data", string(data)))
	}
	return err
}

func DeserializeInput(ctx context.Context, data []byte, out any) error {
	err := json.Unmarshal(data, out)
	if err != nil && logger.ShouldLogDebug() {
		logger.DebugError(ctx, err, "json.Unmarshal")
	}
	return err
}
