package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func TimeoutUnary(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, _ *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (response any, err error) {
		requestCtx, requestCancelCtx := context.WithTimeout(ctx, timeout)
		defer requestCancelCtx()

		return handle(requestCtx, request)
	}
}
