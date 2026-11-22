package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const KeyCallerService = "caller-service"

func CallerServiceFromIncomingContext(ctx context.Context) string {
	values := metadata.ValueFromIncomingContext(ctx, KeyCallerService)
	if len(values) != 1 {
		return ""
	}
	return values[0]
}
