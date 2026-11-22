package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vnafikov/foundation/infrastructure/databases/pg/tracer"
	"github.com/vnafikov/foundation/logger"
)

var Pool *pgxpool.Pool

func CreatePool(ctx context.Context, dsn string) error {
	if Pool != nil {
		conn, err := Pool.Acquire(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(ctx, err, "pgxpool.(*Pool).Acquire")
			}
			return err
		}

		conn.Release()
		return nil
	}

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, err, "pgxpool.ParseConfig")

		return err
	}

	conf.ConnConfig.Tracer = new(tracer.Tracer)
	Pool, err = pgxpool.NewWithConfig(ctx, conf)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err, "pgxpool.NewWithConfig")
	}
	return err
}

func ClosePool() {
	if Pool == nil {
		return
	}

	Pool.Close()
	Pool = nil
}
