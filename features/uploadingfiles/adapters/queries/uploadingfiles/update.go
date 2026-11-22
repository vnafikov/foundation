package uploadingfiles

import (
	"context"

	"github.com/vnafikov/foundation/features/uploadingfiles/core/models"
	"github.com/vnafikov/foundation/features/uploadingfiles/core/queries/uploadingfileparameters"
)

func (q Queries) LeaseExpiredBatch(
	ctx context.Context,
	parameters *uploadingfileparameters.LeaseExpiredBatch,
) ([]*models.UploadingFile, error) {
	const sql = `
WITH expired AS (
	SELECT path
	FROM uploading_files
	WHERE created_at < now() - $1::INTERVAL
		AND (
			leased_at IS NULL
			OR leased_at < now() - $2::INTERVAL
		)
	ORDER BY created_at
	LIMIT $3
	FOR UPDATE SKIP LOCKED
)
UPDATE uploading_files uf
SET lease_token = $4, leased_at = now()
WHERE uf.path = ANY(
	ARRAY(
		SELECT path
		FROM expired
	)
)
RETURNING ` + models.UploadingFileDefaultColumns

	return q.q.QueryRows(ctx, sql, parameters.TTLInterval, parameters.LeaseInterval, parameters.BatchSize, parameters.LeaseToken)
}
