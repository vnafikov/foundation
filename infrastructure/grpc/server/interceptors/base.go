package interceptors

import (
	"time"

	"google.golang.org/grpc"
)

func BaseUnary(requestTimeout time.Duration, grpcError func(error) error) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		RecoveryUnary,
		TraceUnary,
		TimeoutUnary(requestTimeout),
		ErrorUnary(grpcError),
		CallerServiceUnary,
	}
}

func BaseStream(grpcError func(error) error) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		RecoveryStream,
		TraceStream,
		ErrorStream(grpcError),
		CallerServiceStream,
	}
}
