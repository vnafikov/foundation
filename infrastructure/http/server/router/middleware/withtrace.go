package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers/request"
	"github.com/vnafikov/foundation/logger"
	"github.com/vnafikov/foundation/logger/parameters"
)

func withTrace(next func(request.BaseHandlerContext)) func(request.BaseHandlerContext) {
	return func(c request.BaseHandlerContext) {
		if !logger.ShouldLogDebug() {
			next(c)
			return
		}

		startTime := traceStart(c.Ctx, c.Request)
		next(c)
		traceEnd(c.Ctx, c.Writer, c.Request, startTime)
	}
}

func traceStart(ctx context.Context, r *http.Request) (startTime time.Time) {
	startTime = time.Now()

	msg := parameters.HTTPStartMessagePrefix + " " +
		r.Method + " " +
		r.Host + r.RequestURI
	logger.Debug(ctx, msg)

	return startTime
}

func traceEnd(ctx context.Context, w *request.ResponseWriter, r *http.Request, startTime time.Time) {
	msg := parameters.HTTPEndMessagePrefix + " " +
		r.Method + " " +
		strconv.Itoa(w.StatusCode()) + " " +
		r.Host + r.RequestURI
	duration := time.Since(startTime)
	logger.Debug(ctx, msg, slog.Duration("duration", duration))
}
