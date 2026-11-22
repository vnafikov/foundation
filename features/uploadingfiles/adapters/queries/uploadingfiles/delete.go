package uploadingfiles

import "context"

/*
DeleteSQL запрос удаляет запись о загрузке файла.

IMPORTANT: Должно исполняться в транзакции (например, с постановкой задачи) после завершения загрузки файла.
*/
const DeleteSQL = `
DELETE FROM uploading_files
WHERE path = ` // $1.

func (q Queries) Delete(ctx context.Context, path string) error {
	const sql = DeleteSQL + `$1
`

	_, err := q.q.Exec(ctx, sql, path)
	return err
}

func (q Queries) DeleteByLeaseToken(ctx context.Context, leaseToken string) (count int64, err error) {
	const sql = `
DELETE FROM uploading_files
WHERE lease_token = $1
`

	commandTag, err := q.q.Exec(ctx, sql, leaseToken)
	return commandTag.RowsAffected(), err
}
