package backgroundtasks

import (
	"context"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/queries/backgroundtaskparameters"
)

/*
LeaseBatchByRunner арендует для runner пакет независимых фоновых задач или одну последовательную группу.

Выбирает голову по id.
Если у неё нет ordering_key, то арендует до BatchSize независимых фоновых задач.
Иначе арендует по id до OrderedBatchSize фоновых задач того же ordering_key, чтобы один ключ не занял всю ёмкость runner.
Повторными вызовами runner добирает другие группы и независимые задачи до общего BatchSize.
Блокирует строку этой задачи, чтобы конкурентный worker не мог арендовать следующую задачу того же ключа.
Упорядоченную группу арендует, начиная с задачи с минимальным id среди оставшихся задач этого ключа.
Исключает задачи из ExcludedIDs, пока runner не закончил их локальную обработку.
Останавливает набор группы перед отложенной задачей, задачей с действующей арендой или задачей из ExcludedIDs.
Очищает ключ дедупликации при аренде. Ключ не должен восстанавливаться при повторе задачи.
*/
func (q Queries) LeaseBatchByRunner(
	ctx context.Context,
	parameters *backgroundtaskparameters.LeaseBatchByRunner,
) ([]*models.BackgroundTask, error) {
	const sql = `
WITH head AS (
	SELECT bt.id, bt.ordering_key
	FROM background_tasks bt
	WHERE bt.runner = $1
		AND bt.available_at <= now()
		AND bt.id != ALL($2)
		AND (
			bt.leased_at IS NULL
			OR bt.leased_at <= now() - $3::INTERVAL
		)
		AND (
			bt.ordering_key IS NULL
			OR NOT EXISTS (
				SELECT TRUE
				FROM background_tasks earlier_task
				WHERE earlier_task.runner = bt.runner
					AND earlier_task.ordering_key = bt.ordering_key
					AND earlier_task.id < bt.id
			)
		)
	ORDER BY bt.id
	LIMIT 1
	FOR UPDATE OF bt SKIP LOCKED
),
unordered_tasks AS (
	SELECT t.id
	FROM head h
	CROSS JOIN LATERAL (
		SELECT bt.id
		FROM background_tasks bt
		WHERE h.ordering_key IS NULL
			AND bt.runner = $1
			AND bt.ordering_key IS NULL
			AND bt.available_at <= now()
			AND bt.id != ALL($2)
			AND (
				bt.leased_at IS NULL
				OR bt.leased_at <= now() - $3::INTERVAL
			)
		ORDER BY bt.id
		LIMIT $4
		FOR UPDATE OF bt SKIP LOCKED
	) t
),
ordered_tasks AS (
	SELECT t.id
	FROM head h
	CROSS JOIN LATERAL (
		SELECT bt.id
		FROM background_tasks bt
		WHERE h.ordering_key IS NOT NULL
			AND bt.runner = $1
			AND bt.ordering_key = h.ordering_key
			AND bt.available_at <= now()
			AND bt.id != ALL($2)
			AND (
				bt.leased_at IS NULL
				OR bt.leased_at <= now() - $3::INTERVAL
			)
			AND NOT EXISTS (
				SELECT TRUE
				FROM background_tasks earlier_task
				WHERE earlier_task.runner = bt.runner
					AND earlier_task.ordering_key = bt.ordering_key
					AND earlier_task.id < bt.id
					AND (
						earlier_task.available_at > now()
						OR earlier_task.leased_at > now() - $3::INTERVAL
						OR earlier_task.id = ANY($2)
					)
			)
		ORDER BY bt.id
		LIMIT $5
		FOR UPDATE OF bt SKIP LOCKED
	) t
),
tasks AS (
	SELECT id
	FROM unordered_tasks
	UNION ALL
	SELECT id
	FROM ordered_tasks
),
leased_tasks AS (
	UPDATE background_tasks bt
	SET deduplication_key = NULL,
		leased_by_worker_id = $6,
		leased_at = now()
	WHERE bt.id = ANY(
		ARRAY(
			SELECT id
			FROM tasks
		)
	)
	RETURNING ` + models.BackgroundTaskDefaultColumns + `
)
SELECT ` + models.BackgroundTaskDefaultColumns + `
FROM leased_tasks bt
ORDER BY bt.id
`

	return q.q.QueryRows(
		ctx,
		sql,
		parameters.Runner,
		parameters.ExcludedIDs,
		parameters.LeaseInterval,
		parameters.BatchSize,
		parameters.OrderedBatchSize,
		parameters.WorkerID,
	)
}

func (q Queries) UpdateLeases(ctx context.Context, parameters *backgroundtaskparameters.UpdateLeases) error {
	const sql = `
UPDATE background_tasks
SET leased_at = now()
WHERE id = ANY($1)
	AND leased_by_worker_id = $2
`

	_, err := q.q.Exec(ctx, sql, parameters.IDs, parameters.LeasedByWorkerID)
	return err
}

func (q Queries) ResetLeases(ctx context.Context, leasedByWorkerID int) error {
	const sql = `
UPDATE background_tasks
SET leased_by_worker_id = NULL, leased_at = NULL
WHERE leased_by_worker_id = $1
`

	_, err := q.q.Exec(ctx, sql, leasedByWorkerID)
	return err
}

func (q Queries) ResetLeasesByIDs(ctx context.Context, parameters *backgroundtaskparameters.ResetLeasesByIDs) error {
	const sql = `
UPDATE background_tasks
SET leased_by_worker_id = NULL, leased_at = NULL
WHERE id = ANY($1)
	AND leased_by_worker_id = $2
`

	_, err := q.q.Exec(ctx, sql, parameters.IDs, parameters.LeasedByWorkerID)
	return err
}

func (q Queries) Reschedule(ctx context.Context, parameters *backgroundtaskparameters.Reschedule) error {
	const sql = `
UPDATE background_tasks
SET available_at = $1,
	leased_by_worker_id = NULL,
	leased_at = NULL
WHERE id = $2
	AND leased_by_worker_id = $3
`

	_, err := q.q.Exec(ctx, sql, parameters.AvailableAt, parameters.ID, parameters.LeasedByWorkerID)
	return err
}
