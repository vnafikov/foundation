package interceptors

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vnafikov/foundation/infrastructure/grpc/metadata"
	"github.com/vnafikov/foundation/logger"
	"github.com/vnafikov/foundation/logger/parameters"
)

func TraceUnary(ctx context.Context, request any, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (response any, err error) {
	if !logger.ShouldLogDebug() {
		return handle(ctx, request)
	}

	attrs := traceAttrs(ctx)
	startTime := traceStart(ctx, info.FullMethod, attrs)
	response, err = handle(ctx, request)
	traceEnd(ctx, info.FullMethod, err, attrs, startTime)
	return response, err
}

func traceAttrs(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr
	callerService := metadata.CallerServiceFromIncomingContext(ctx)
	if callerService != "" {
		attrs = append(attrs, slog.String("caller_service", callerService))
	}
	return attrs
}

func traceStart(ctx context.Context, fullMethod string, attrs []slog.Attr) time.Time {
	startTime := time.Now()

	msg := parameters.GRPCStartMessagePrefix + " " +
		fullMethod
	logger.Debug(ctx, msg, attrs...)

	return startTime
}

func traceEnd(ctx context.Context, fullMethod string, err error, attrs []slog.Attr, startTime time.Time) {
	msg := parameters.GRPCEndMessagePrefix + " " +
		statusCode(err) + " " +
		fullMethod
	duration := time.Since(startTime)
	attrs = append(attrs, slog.Duration("duration", duration))
	logger.Debug(ctx, msg, attrs...)
}

func statusCode(err error) string {
	var code codes.Code
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		code = status.FromContextError(err).Code()
	} else {
		code = status.Code(err)
	}
	return code.String()
}

func TraceStream(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handle grpc.StreamHandler) error {
	if !logger.ShouldLogDebug() {
		return handle(srv, stream)
	}

	ctx := stream.Context()
	attrs := traceAttrs(ctx)
	startTime := traceStart(ctx, info.FullMethod, attrs)
	err := handle(srv, stream)
	traceEnd(ctx, info.FullMethod, err, attrs, startTime)
	return err
}
