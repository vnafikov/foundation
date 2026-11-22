package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vnafikov/foundation/infrastructure/grpc/metadata"
)

func CallerServiceUnary(ctx context.Context, request any, _ *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (response any, err error) {
	if metadata.CallerServiceFromIncomingContext(ctx) == "" {
		return nil, status.Error(codes.Unauthenticated, "")
	}
	return handle(ctx, request)
}

func CallerServiceStream(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handle grpc.StreamHandler) error {
	if metadata.CallerServiceFromIncomingContext(stream.Context()) == "" {
		return status.Error(codes.Unauthenticated, "")
	}
	return handle(srv, stream)
}
