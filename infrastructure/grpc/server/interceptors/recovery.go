package interceptors

import (
	"context"
	"log/slog"

	"github.com/vnafikov/common/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vnafikov/foundation/logger"
)

const skipStackFramesCount = 5

func RecoveryUnary(ctx context.Context, request any, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (response any, err error) {
	defer func() {
		if e := recover(); e != nil {
			logPanic(ctx, e, info.FullMethod)

			err = status.Error(codes.Internal, "")
		}
	}()

	return handle(ctx, request)
}

func logPanic(ctx context.Context, err any, fullMethod string) {
	msg := stacktrace.PanicMessage(err, skipStackFramesCount)
	logger.ErrorMsgWithoutStack(ctx, msg, slog.String("grpc_method", fullMethod))
}

func RecoveryStream(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handle grpc.StreamHandler) (err error) {
	defer func() {
		if e := recover(); e != nil {
			logPanic(stream.Context(), e, info.FullMethod)

			err = status.Error(codes.Internal, "")
		}
	}()

	return handle(srv, stream)
}
