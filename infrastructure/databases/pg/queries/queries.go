package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/vnafikov/foundation/logger"
)

type Queries[T Hydratable[T]] struct {
	BaseQueries
}

/*
QueryRow возвращает одну запись.

Если запись не найдена, то вернёт nil без ошибки.
*/
func (q Queries[T]) QueryRow(ctx context.Context, sql string, args []any) (*T, error) {
	row := q.Conn.QueryRow(ctx, sql, args...)
	entity, err := collectRow[T](row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		if shouldLogError(err) {
			logger.Error(ctx, err, "queries.collectRow")
		}
	}
	return entity, err
}

func collectRow[T Hydratable[T]](row pgx.Row) (*T, error) {
	var zero T
	entity, fields := zero.Fields()
	if err := row.Scan(fields...); err != nil {
		return nil, err
	}
	return entity, nil
}

/*
QueryRowStrict возвращает одну запись.

Если запись не найдена, то вернёт nil с ошибкой [pgx.ErrNoRows].
*/
func (q Queries[T]) QueryRowStrict(ctx context.Context, sql string, args []any) (*T, error) {
	row := q.Conn.QueryRow(ctx, sql, args...)
	entity, err := collectRow[T](row)
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "queries.collectRow")
	}
	return entity, err
}

/*
QueryRows возвращает срез записей.

Если записи не найдены, то вернёт пустой срез без ошибки.
*/
func (q Queries[T]) QueryRows(ctx context.Context, sql string, args []any) ([]*T, error) {
	rows, err := q.Conn.Query(ctx, sql, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(ctx, err, "queries.(Querier).Query")
		}
		return nil, err
	}

	entities, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*T, error) { return collectRow[T](row) })
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "pgx.CollectRows")
	}
	return entities, err
}
