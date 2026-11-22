package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/vnafikov/foundation/logger"
)

type ConnQueries struct {
	conn *pgx.Conn
}

func (cq ConnQueries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	commandTag, err := cq.conn.Exec(ctx, sql, args...)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err, "pgx.(*Conn).Exec")
	}
	return commandTag, err
}

func (cq ConnQueries) WaitForNotification(ctx context.Context) (*pgconn.Notification, error) {
	notification, err := cq.conn.WaitForNotification(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error(ctx, err, "pgx.(*Conn).WaitForNotification")
	}
	return notification, err
}

func (cq ConnQueries) Listen(ctx context.Context, channel string) error {
	sql := "LISTEN " + channel

	_, err := cq.Exec(ctx, sql)
	return err
}

func NewConnQueries(conn *pgx.Conn) ConnQueries {
	return ConnQueries{conn: conn}
}
