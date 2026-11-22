package client

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/vnafikov/foundation/logger"
)

type Conns map[string]*grpc.ClientConn

func (conns Conns) ByName(ctx context.Context, name string) (*grpc.ClientConn, error) {
	conn, ok := conns[name]
	if !ok {
		logger.Error(ctx, ErrConnNotFound, "client.Conns", slog.String("name", name))

		return nil, ErrConnNotFound
	}
	return conn, nil
}
