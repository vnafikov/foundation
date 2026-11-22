package localfs

import (
	"context"

	"github.com/vnafikov/foundation/infrastructure/storages/localfs"
)

type DeleteFilesIfExists struct {
	rootPath string
}

func (d *DeleteFilesIfExists) Handle(ctx context.Context, paths []string) error {
	return localfs.DeleteFilesIfExistsInRoot(ctx, d.rootPath, paths)
}

func NewDeleteFilesIfExists(rootPath string) DeleteFilesIfExists {
	return DeleteFilesIfExists{rootPath: rootPath}
}
