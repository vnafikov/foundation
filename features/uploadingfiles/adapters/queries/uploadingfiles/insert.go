package uploadingfiles

import "context"

func (q Queries) Insert(ctx context.Context, path string) error {
	const sql = `
INSERT INTO uploading_files (path)
VALUES ($1)
`

	_, err := q.q.Exec(ctx, sql, path)
	return err
}
