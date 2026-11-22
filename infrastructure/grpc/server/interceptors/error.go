package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

func ErrorUnary(grpcError func(error) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, _ *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (response any, err error) {
		response, err = handle(ctx, request)
		if err != nil {
			err = grpcError(err)
		}
		return response, err
	}
}

func ErrorStream(grpcError func(error) error) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handle grpc.StreamHandler) error {
		err := handle(srv, stream)
		if err != nil {
			err = grpcError(err)
		}
		return err
	}
}
