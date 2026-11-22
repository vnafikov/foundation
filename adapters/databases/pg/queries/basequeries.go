package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	paginationparameters "github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
)

type BaseQueries struct {
	q queries.BaseQueries
}

func (q *BaseQueries) UseConn(conn queries.Querier) {
	q.q.Conn = conn
}

func (q BaseQueries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	commandTag, err := q.q.Exec(ctx, sql, args)
	return commandTag, contractError[any](err)
}

func (q BaseQueries) QueryRowValue(ctx context.Context, value any, sql string, args ...any) (exists bool, err error) {
	exists, err = q.q.QueryRowValue(ctx, value, sql, args)
	return exists, contractError[any](err)
}

func (q BaseQueries) QueryRowValueStrict(ctx context.Context, value any, sql string, args ...any) error {
	err := q.q.QueryRowValueStrict(ctx, value, sql, args)
	return contractError[any](err)
}

func (q BaseQueries) QueryRowFields[T any](ctx context.Context, fields func(*T) []any, sql string, args ...any) (*T, error) {
	entity, err := q.q.QueryRowFields(ctx, fields, sql, args)
	return entity, contractError[any](err)
}

func (q BaseQueries) QueryRowFieldsStrict[T any](ctx context.Context, fields func(*T) []any, sql string, args ...any) (*T, error) {
	entity, err := q.q.QueryRowFieldsStrict(ctx, fields, sql, args)
	return entity, contractError[any](err)
}

func (q BaseQueries) QueryRowsValue[T any](ctx context.Context, sql string, args ...any) ([]T, error) {
	values, err := q.q.QueryRowsValue[T](ctx, sql, args)
	return values, contractError[any](err)
}

func (q BaseQueries) QueryRowsFields[T any](ctx context.Context, fields func(*T) []any, sql string, args ...any) ([]*T, error) {
	entities, err := q.q.QueryRowsFields(ctx, fields, sql, args)
	return entities, contractError[any](err)
}

func (q BaseQueries) Exists(ctx context.Context, sql string, args ...any) (bool, error) {
	exists, err := q.q.Exists(ctx, sql, args)
	return exists, contractError[any](err)
}

func (q BaseQueries) SelectPageFieldsByCursor[T any](
	ctx context.Context,
	fields func(*T) []any,
	parameters *SelectPageByCursorParameters[T],
) ([]*T, paginationparameters.CursorMeta, error) {
	entities, cursorMeta, err := selectPageByCursor(ctx, parameters, func(ctx context.Context, sql string, args []any) ([]*T, error) {
		return q.q.QueryRowsFields(ctx, fields, sql, args)
	})
	return entities, cursorMeta, contractError[any](err)
}

func (q BaseQueries) SelectPageFieldsByOffset[T any](
	ctx context.Context,
	fields func(*T) []any,
	parameters *SelectPageByOffsetParameters,
) ([]*T, error) {
	entities, err := selectPageByOffset(ctx, parameters, func(ctx context.Context, sql string, args []any) ([]*T, error) {
		return q.q.QueryRowsFields(ctx, fields, sql, args)
	})
	return entities, contractError[any](err)
}
