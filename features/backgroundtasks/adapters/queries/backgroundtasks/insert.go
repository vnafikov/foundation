package backgroundtasks

import (
	"context"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/queries/backgroundtaskparameters"
)

const (
	/*
		InsertSQL запрос создаёт запись с фоновой задачей.

		IMPORTANT: Изменение данных и постановка связанной задачи должны выполняться в одной транзакции.
	*/
	InsertSQL = `
INSERT INTO background_tasks (name, runner, payload)` // VALUES ($1, $2, $3).
	/*
		InsertOrderedSQL запрос создаёт запись с последовательной фоновой задачей.

		IMPORTANT: Изменение данных и постановка связанной задачи должны выполняться в одной транзакции.
		Использовать, если общим механизмом сериализации для пары (runner, ordering_key) является внешняя блокировка,
		удерживаемая до завершения транзакции.
	*/
	InsertOrderedSQL = `
INSERT INTO background_tasks (name, runner, ordering_key, payload)` // VALUES ($1, $2, $3, $4).
	/*
		InsertDeduplicatedSQL запрос создаёт запись с последовательной дедуплицированной фоновой задачей.

		IMPORTANT: Изменение данных и постановка связанной задачи должны выполняться в одной транзакции.
		Использовать, если общим механизмом сериализации для пары (runner, ordering_key) является внешняя блокировка,
		удерживаемая до завершения транзакции.
	*/
	InsertDeduplicatedSQL = `
INSERT INTO background_tasks (name, runner, ordering_key, deduplication_key, available_at)` // VALUES ($1, $2, $3, $4, $5).
)

/*
OrderingLockSQL получает блокировку для постановки последовательной фоновой задачи.

Transaction-level advisory lock сериализует постановку только в пределах пары (runner, ordering_key).
Вставка должна использовать FROM ordering_lock, тогда ID задачи генерируется только после получения блокировки.

IMPORTANT: Изменение данных и постановка связанной задачи должны выполняться в одной транзакции.
Использовать, если общим механизмом сериализации для пары (runner, ordering_key) является создаваемый этим запросом advisory lock.
*/
func OrderingLockSQL(runner, orderingKey, from string) string {
	if from != "" {
		from = `
	FROM ` + from
	}
	return `ordering_lock AS MATERIALIZED (
	SELECT pg_advisory_xact_lock(
		hashtextextended(jsonb_build_array(` + runner + `::background_task_runner_type, ` + orderingKey + `::TEXT)::TEXT, 0)
	)` + from + `
)`
}

/*
OrderingLocksSQL получает блокировки для постановки последовательных фоновых задач.

Transaction-level advisory locks сериализуют постановку только в пределах соответствующих пар (runner, ordering_key).
Блокировки получает в порядке ordering_key.
Вставка должна использовать CROSS JOIN acquired_ordering_locks, тогда ID задач генерируются только после получения всех блокировок.

IMPORTANT: Изменение данных и постановка связанной задачи должны выполняться в одной транзакции.
Использовать, если общим механизмом сериализации для пары (runner, ordering_key) является создаваемый этим запросом advisory lock.
*/
func OrderingLocksSQL(runner, from string) string {
	return `ordering_locks AS MATERIALIZED (
	SELECT pg_advisory_xact_lock(hashtextextended(jsonb_build_array(` + runner + `::background_task_runner_type, ordering_key)::TEXT, 0))
	FROM ` + from + `
	ORDER BY ordering_key
),
acquired_ordering_locks AS (
	SELECT count(*)
	FROM ordering_locks
)`
}

/*
Insert создаёт фоновую задачу.
При наличии ключа дедупликации пропускает дубликат с теми же исполнителем и ключом.
Для дедупликации обязателен ключ последовательного выполнения.
*/
func (q Queries) Insert(ctx context.Context, parameters *backgroundtaskparameters.Insert) error {
	if parameters.OrderingKey.IsValid {
		sql := `
WITH ` + OrderingLockSQL("$2", "$3", "") + `
INSERT INTO background_tasks (name, runner, ordering_key, deduplication_key, payload, available_at)
SELECT $1, $2, $3, $4, $5, COALESCE($6, now())
FROM ordering_lock
ON CONFLICT (runner, deduplication_key) WHERE deduplication_key IS NOT NULL
DO NOTHING
`

		_, err := q.q.Exec(
			ctx,
			sql,
			parameters.Definition.Name,
			parameters.Definition.Runner,
			parameters.OrderingKey.String,
			parameters.DeduplicationKey,
			parameters.Payload,
			parameters.AvailableAt,
		)
		return err
	}

	const sql = `
INSERT INTO background_tasks (name, runner, payload, available_at)
VALUES ($1, $2, $3, COALESCE($4, now()))
`

	_, err := q.q.Exec(
		ctx,
		sql,
		parameters.Definition.Name,
		parameters.Definition.Runner,
		parameters.Payload,
		parameters.AvailableAt,
	)
	return err
}
