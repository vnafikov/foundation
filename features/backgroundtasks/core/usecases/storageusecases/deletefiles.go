package storageusecases

import (
	"context"

	"github.com/vnafikov/foundation/adapters/storages/localfs"
)

type DeleteFiles struct {
	deleteFiles localfs.DeleteFilesIfExists
}

// Handle удаляет файлы по указанным путям (пустые папки останутся).
func (useCase *DeleteFiles) Handle(ctx context.Context, paths []string) error {
	return useCase.deleteFiles.Handle(ctx, paths)
}

func NewDeleteFiles(deleteFiles localfs.DeleteFilesIfExists) DeleteFiles {
	return DeleteFiles{deleteFiles: deleteFiles}
}
