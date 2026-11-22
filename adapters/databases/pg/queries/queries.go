package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	paginationparameters "github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
)

type Queries[T queries.Hydratable[T]] struct {
	q queries.Queries[T]
}

func (q *Queries[T]) UseConn(conn queries.Querier) {
	q.q.Conn = conn
}

func (q Queries[T]) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	commandTag, err := q.q.Exec(ctx, sql, args)
	return commandTag, contractError[T](err)
}

func (q Queries[T]) QueryRowValue(ctx context.Context, value any, sql string, args ...any) (exists bool, err error) {
	exists, err = q.q.QueryRowValue(ctx, value, sql, args)
	return exists, contractError[T](err)
}

func (q Queries[T]) QueryRowValueStrict(ctx context.Context, value any, sql string, args ...any) error {
	err := q.q.QueryRowValueStrict(ctx, value, sql, args)
	return contractError[T](err)
}

func (q Queries[T]) QueryRowFields[R any](ctx context.Context, fields func(*R) []any, sql string, args ...any) (*R, error) {
	entity, err := q.q.QueryRowFields(ctx, fields, sql, args)
	return entity, contractError[T](err)
}

func (q Queries[T]) QueryRowFieldsStrict[R any](ctx context.Context, fields func(*R) []any, sql string, args ...any) (*R, error) {
	entity, err := q.q.QueryRowFieldsStrict(ctx, fields, sql, args)
	return entity, contractError[T](err)
}

func (q Queries[T]) QueryRowsValue[R any](ctx context.Context, sql string, args ...any) ([]R, error) {
	values, err := q.q.QueryRowsValue[R](ctx, sql, args)
	return values, contractError[T](err)
}

func (q Queries[T]) QueryRowsFields[R any](ctx context.Context, fields func(*R) []any, sql string, args ...any) ([]*R, error) {
	entities, err := q.q.QueryRowsFields(ctx, fields, sql, args)
	return entities, contractError[T](err)
}

func (q Queries[T]) Exists(ctx context.Context, sql string, args ...any) (bool, error) {
	exists, err := q.q.Exists(ctx, sql, args)
	return exists, contractError[T](err)
}

func (q Queries[T]) QueryRow(ctx context.Context, sql string, args ...any) (*T, error) {
	entity, err := q.q.QueryRow(ctx, sql, args)
	return entity, contractError[T](err)
}

func (q Queries[T]) QueryRowStrict(ctx context.Context, sql string, args ...any) (*T, error) {
	entity, err := q.q.QueryRowStrict(ctx, sql, args)
	return entity, contractError[T](err)
}

func (q Queries[T]) QueryRows(ctx context.Context, sql string, args ...any) ([]*T, error) {
	entities, err := q.q.QueryRows(ctx, sql, args)
	return entities, contractError[T](err)
}

func (q Queries[T]) SelectPageRowsByCursor(
	ctx context.Context,
	parameters *SelectPageByCursorParameters[T],
) ([]*T, paginationparameters.CursorMeta, error) {
	entities, cursorMeta, err := selectPageByCursor(ctx, parameters, q.q.QueryRows)
	return entities, cursorMeta, contractError[T](err)
}

func (q Queries[T]) SelectPageFieldsByCursor[R any](
	ctx context.Context,
	fields func(*R) []any,
	parameters *SelectPageByCursorParameters[R],
) ([]*R, paginationparameters.CursorMeta, error) {
	entities, cursorMeta, err := selectPageByCursor(ctx, parameters, func(ctx context.Context, sql string, args []any) ([]*R, error) {
		return q.q.QueryRowsFields(ctx, fields, sql, args)
	})
	return entities, cursorMeta, contractError[T](err)
}

func (q Queries[T]) SelectPageRowsByOffset(ctx context.Context, parameters *SelectPageByOffsetParameters) ([]*T, error) {
	entities, err := selectPageByOffset(ctx, parameters, q.q.QueryRows)
	return entities, contractError[T](err)
}

func (q Queries[T]) SelectPageFieldsByOffset[R any](
	ctx context.Context,
	fields func(*R) []any,
	parameters *SelectPageByOffsetParameters,
) ([]*R, error) {
	entities, err := selectPageByOffset(ctx, parameters, func(ctx context.Context, sql string, args []any) ([]*R, error) {
		return q.q.QueryRowsFields(ctx, fields, sql, args)
	})
	return entities, contractError[T](err)
}
