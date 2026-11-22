package tasks

import (
	"context"
	"log/slog"

	"github.com/vnafikov/foundation/adapters/storages/localfs"
	"github.com/vnafikov/foundation/features/uploadingfiles/adapters/queries/uploadingfiles"
	"github.com/vnafikov/foundation/features/uploadingfiles/core/usecases/uploadingfileusecases"
	"github.com/vnafikov/foundation/infrastructure/databases/pg"
)

func DeleteExpiredUploadingFiles(
	ctx context.Context,
	rootPath string,
	params *uploadingfileusecases.DeleteExpiredParameters,
) (logAttrs []slog.Attr, err error) {
	uploadingFileQueries := uploadingfiles.New(pg.Pool)
	deleteFilesIfExists := localfs.NewDeleteFilesIfExists(rootPath)
	useCase := uploadingfileusecases.NewDeleteExpired(uploadingFileQueries, deleteFilesIfExists)
	count, err := useCase.Handle(ctx, params)
	logAttrs = []slog.Attr{slog.Int64("deleted_count", count)}
	return logAttrs, err
}
