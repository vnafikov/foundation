package backgroundtasks

import (
	"context"

	"github.com/vnafikov/common/binarycodec"
	"github.com/vnafikov/common/types/nullable"

	"github.com/vnafikov/foundation/adapters/databases/pg/builders"
	"github.com/vnafikov/foundation/adapters/databases/pg/queries"
	"github.com/vnafikov/foundation/core/parameters/page"
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/queries/backgroundtaskparameters"
)

const selectSQL = `
SELECT ` + models.BackgroundTaskDefaultColumns + `
FROM background_tasks bt`

// SelectAvailableAtByRunner возвращает ближайший срок доступности с учётом аренды и порядка задач.
func (q Queries) SelectAvailableAtByRunner(
	ctx context.Context,
	parameters *backgroundtaskparameters.SelectAvailableAtByRunner,
) (availableAt nullable.Time, err error) {
	const sql = `
SELECT min(available_at)
FROM (
	(
		SELECT available_at
		FROM background_tasks
		WHERE runner = $1
			AND ordering_key IS NULL
			AND leased_at IS NULL
			AND id != ALL($2)
		ORDER BY available_at
		LIMIT 1
	)
	UNION ALL
	SELECT min(GREATEST(available_at, leased_at + $3))
	FROM background_tasks
	WHERE runner = $1
		AND ordering_key IS NULL
		AND leased_at IS NOT NULL
		AND id != ALL($2)
	UNION ALL
	SELECT min(GREATEST(available_at, leased_at + $3))
	FROM (
		SELECT DISTINCT ON (ordering_key) id, available_at, leased_at
		FROM background_tasks
		WHERE runner = $1
			AND ordering_key IS NOT NULL
		ORDER BY ordering_key, id
	)
	WHERE id != ALL($2)
)
`

	err = q.q.QueryRowValueStrict(ctx, &availableAt, sql, parameters.Runner, parameters.ExcludedIDs, parameters.LeaseInterval)
	return availableAt, err
}

func (q Queries) SelectPage(ctx context.Context, pageParameters *page.Parameters) ([]*models.BackgroundTask, pagination.CursorMeta, error) {
	var id int64
	return q.q.SelectPageRowsByCursor(ctx, &queries.SelectPageByCursorParameters[models.BackgroundTask]{
		CursorPagination: pageParameters.Pagination.Cursor,
		ReadCursor: func(reader binarycodec.Reader) bool {
			var err error
			id, err = reader.ReadInt64()
			return err == nil
		},
		CursorCapacity: binarycodec.SizeInt64,
		WriteCursor: func(writer *binarycodec.Writer, entity *models.BackgroundTask) error {
			writer.WriteInt64(entity.ID)
			return nil
		},
		SelectFirstPage: func(limit int) (string, []any) {
			whereClause, values := builders.BuildWhereClause(pageParameters.Filtering.Conditions, 1)
			values[0] = limit

			sql := selectSQL + whereClause + `
ORDER BY id
LIMIT $1 + 1
`

			return sql, values
		},
		SelectNextPage: func(limit int) (string, []any) {
			whereCondition, and, values := builders.BuildWhereCondition(pageParameters.Filtering.Conditions, 2)
			values[0] = id
			values[1] = limit

			sql := selectSQL + `
WHERE ` + whereCondition + and + `id > $1
ORDER BY id
LIMIT $2 + 1
`

			return sql, values
		},
		SelectPreviousPage: func(limit int) (string, []any) {
			whereCondition, and, values := builders.BuildWhereCondition(pageParameters.Filtering.Conditions, 2)
			values[0] = id
			values[1] = limit

			sql := `
SELECT *
FROM (` + selectSQL + `
	WHERE ` + whereCondition + and + `id < $1
	ORDER BY id DESC
	LIMIT $2 + 1
)
ORDER BY id
`

			return sql, values
		},
		SelectLastPage: func(limit int) (string, []any) {
			whereClause, values := builders.BuildWhereClause(pageParameters.Filtering.Conditions, 1)
			values[0] = limit

			sql := `
SELECT *
FROM (` + selectSQL + whereClause + `
	ORDER BY id DESC
	LIMIT $1 + 1
)
ORDER BY id
`

			return sql, values
		},
	})
}

func (q Queries) SelectCount(ctx context.Context, filteringParameters *filtering.Parameters) (count int64, err error) {
	whereClause, values := builders.BuildWhereClause(filteringParameters.Conditions, 0)

	sql := `
SELECT count(*)
FROM background_tasks bt` + whereClause

	err = q.q.QueryRowValueStrict(ctx, &count, sql, values...)
	return count, err
}
