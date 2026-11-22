package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/vnafikov/foundation/logger"
)

type BaseQueries struct {
	Conn Querier
}

func (q BaseQueries) Exec(ctx context.Context, sql string, args []any) (pgconn.CommandTag, error) {
	commandTag, err := q.Conn.Exec(ctx, sql, args...)
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "queries.(Querier).Exec")
	}
	return commandTag, err
}

func shouldLogError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}

	if pgError, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgError.Code {
		case ErrorCodeUniqueViolation, ErrorCodeCheckViolation, ErrorCodeExclusionViolation:
			return false
		}
	}
	return true
}

/*
QueryRowValue заполняет одно значение.

Если запись не найдена, то вернёт false без ошибки.

value - указатель на переменную.
*/
func (q BaseQueries) QueryRowValue(ctx context.Context, value any, sql string, args []any) (exists bool, err error) {
	row := q.Conn.QueryRow(ctx, sql, args...)
	if err := row.Scan(value); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		if shouldLogError(err) {
			logger.Error(ctx, err, "pgx.(Row).Scan")
		}
		return false, err
	}
	return true, nil
}

/*
QueryRowValueStrict заполняет одно значение.

Если запись не найдена, то вернёт ошибку [pgx.ErrNoRows].

value - указатель на переменную.
*/
func (q BaseQueries) QueryRowValueStrict(ctx context.Context, value any, sql string, args []any) error {
	row := q.Conn.QueryRow(ctx, sql, args...)
	err := row.Scan(value)
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "pgx.(Row).Scan")
	}
	return err
}

/*
QueryRowFields возвращает одну запись с указанными полями.

Если запись не найдена, то вернёт nil без ошибки.

fields - метод должен возвращать срез указателей на необходимые поля структуры.
*/
func (q BaseQueries) QueryRowFields[T any](ctx context.Context, fields func(*T) []any, sql string, args []any) (*T, error) {
	row := q.Conn.QueryRow(ctx, sql, args...)
	entity, err := collectRowFields[T](row, fields)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		if shouldLogError(err) {
			logger.Error(ctx, err, "queries.collectRowFields")
		}
	}
	return entity, err
}

func collectRowFields[T any](row pgx.Row, fields func(*T) []any) (*T, error) {
	entity := new(T)
	if err := row.Scan(fields(entity)...); err != nil {
		return nil, err
	}
	return entity, nil
}

/*
QueryRowFieldsStrict возвращает одну запись с указанными полями.

Если запись не найдена, то вернёт nil с ошибкой [pgx.ErrNoRows].

fields - метод должен возвращать срез указателей на необходимые поля структуры.
*/
func (q BaseQueries) QueryRowFieldsStrict[T any](ctx context.Context, fields func(*T) []any, sql string, args []any) (*T, error) {
	row := q.Conn.QueryRow(ctx, sql, args...)
	entity, err := collectRowFields[T](row, fields)
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "queries.collectRowFields")
	}
	return entity, err
}

/*
QueryRowsValue возвращает срез значений из одной колонки.

Если записи не найдены, то вернёт пустой срез без ошибки.
*/
func (q BaseQueries) QueryRowsValue[T any](ctx context.Context, sql string, args []any) ([]T, error) {
	rows, err := q.Conn.Query(ctx, sql, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(ctx, err, "queries.(Querier).Query")
		}
		return nil, err
	}

	values, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (T, error) {
		var value T
		err := row.Scan(&value)
		return value, err
	})
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "pgx.CollectRows")
	}
	return values, err
}

/*
QueryRowsFields возвращает срез записей с указанными полями.

Если записи не найдены, то вернёт пустой срез без ошибки.

fields - метод должен возвращать срез указателей на необходимые поля структуры.
*/
func (q BaseQueries) QueryRowsFields[T any](ctx context.Context, fields func(*T) []any, sql string, args []any) ([]*T, error) {
	rows, err := q.Conn.Query(ctx, sql, args...)
	if err != nil {
		if shouldLogError(err) {
			logger.Error(ctx, err, "queries.(Querier).Query")
		}
		return nil, err
	}

	entities, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*T, error) { return collectRowFields[T](row, fields) })
	if err != nil && shouldLogError(err) {
		logger.Error(ctx, err, "pgx.CollectRows")
	}
	return entities, err
}

func (q BaseQueries) Exists(ctx context.Context, sql string, args []any) (bool, error) {
	var exists bool
	return q.QueryRowValue(ctx, &exists, sql, args)
}
