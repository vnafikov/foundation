package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

/*
TimeoutUnary задаёт timeout для unary-запроса, если в контексте его ещё нет.

Для частного ограничения времени use case должен передать контекст, созданный через context.WithTimeout или context.WithDeadline.
*/
func TimeoutUnary(defaultTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		request, reply any,
		conn *grpc.ClientConn,
		invoke grpc.UnaryInvoker,
		options ...grpc.CallOption,
	) error {
		if _, ok := ctx.Deadline(); ok {
			return invoke(ctx, method, request, reply, conn, options...)
		}

		requestCtx, requestCancelCtx := context.WithTimeout(ctx, defaultTimeout)
		defer requestCancelCtx()

		return invoke(requestCtx, method, request, reply, conn, options...)
	}
}
