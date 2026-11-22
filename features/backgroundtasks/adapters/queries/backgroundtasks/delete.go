package backgroundtasks

import "context"

/*
DeleteSQL запрос удаляет запись с фоновой задачей.

IMPORTANT: Изменение данных и удаление связанной задачи должны выполняться в одной транзакции.
*/
const DeleteSQL = `
DELETE FROM background_tasks
WHERE id = ` // $1.

func (q Queries) Delete(ctx context.Context, id int64) error {
	const sql = DeleteSQL + `$1
`

	_, err := q.q.Exec(ctx, sql, id)
	return err
}
