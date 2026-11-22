package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	meta "github.com/vnafikov/foundation/infrastructure/grpc/metadata"
)

func CallerServiceUnary(name string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		request, reply any,
		conn *grpc.ClientConn,
		invoke grpc.UnaryInvoker,
		options ...grpc.CallOption,
	) error {
		requestCtx := contextWithCallerService(ctx, name)
		return invoke(requestCtx, method, request, reply, conn, options...)
	}
}

func contextWithCallerService(ctx context.Context, name string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md.Set(meta.KeyCallerService, name)
	return metadata.NewOutgoingContext(ctx, md)
}

func CallerServiceStream(name string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		description *grpc.StreamDesc,
		conn *grpc.ClientConn,
		method string,
		stream grpc.Streamer,
		options ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		requestCtx := contextWithCallerService(ctx, name)
		return stream(requestCtx, description, conn, method, options...)
	}
}
