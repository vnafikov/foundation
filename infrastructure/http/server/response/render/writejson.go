package render

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/servererrors"
	"github.com/vnafikov/foundation/json"
	"github.com/vnafikov/foundation/logger"
)

const (
	headerContentType         = "Content-Type"
	headerXContentTypeOptions = "X-Content-Type-Options"
)

func WriteJSON(ctx context.Context, w http.ResponseWriter, code int, data any) {
	b, err := json.Serialize(ctx, data)
	if err != nil {
		servererrors.InternalServerError(w)
		return
	}

	w.Header().Set(headerContentType, "application/json")
	w.Header().Set(headerXContentTypeOptions, "nosniff")
	w.WriteHeader(code)
	if _, err := w.Write(b); err != nil {
		logWriteError(ctx, err, data)
	}
}

func logWriteError(ctx context.Context, err error, data any) {
	if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, err, "http.(ResponseWriter).Write", slog.Any("data", data))
		}
		return
	}

	logger.Error(ctx, err, "http.(ResponseWriter).Write", slog.Any("data", data))
}
