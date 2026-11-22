package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vnafikov/foundation/infrastructure/databases/pg/tracer"
	"github.com/vnafikov/foundation/logger"
)

const closeTimeout = 15 * time.Second // 15 seconds.

func Connect(ctx context.Context, dsn string) (*pgx.Conn, error) {
	conf, err := pgx.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, err, "pgx.ParseConfig")

		return nil, err
	}

	conf.Tracer = new(tracer.Tracer)
	conn, err := pgx.ConnectConfig(ctx, conf)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err, "pgx.ConnectConfig")
	}
	return conn, err
}

func CloseConn(ctx context.Context, conn *pgx.Conn) error {
	closeCtx, closeCancelCtx := context.WithTimeout(context.WithoutCancel(ctx), closeTimeout)
	defer closeCancelCtx()

	err := conn.Close(closeCtx)
	if err != nil {
		logger.Error(closeCtx, err, "pgx.(*Conn).Close")
	}
	return err
}
