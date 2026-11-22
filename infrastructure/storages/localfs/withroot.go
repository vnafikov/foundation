package localfs

import (
	"context"
	"errors"
	"os"

	"github.com/vnafikov/common/filesystem"

	"github.com/vnafikov/foundation/logger"
)

func WithRoot(ctx context.Context, rootPath string, handle func(FS) error) error {
	err := filesystem.WithRoot(rootPath, func(root *os.Root) error {
		return handle(New(Configuration{Root: root}))
	})
	if errors.Is(err, filesystem.ErrWithRoot) {
		logger.Error(ctx, err, "filesystem.WithRoot")
	}
	return err
}

func CreateDirWithParentsInRoot(ctx context.Context, rootPath, dirPath string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.CreateDirWithParents(ctx, dirPath)
	})
}

func CreateDirsWithParentsInRoot(ctx context.Context, rootPath string, dirPaths []string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.CreateDirsWithParents(ctx, dirPaths)
	})
}

func DeleteFileIfExistsInRoot(ctx context.Context, rootPath, path string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.DeleteFileIfExists(ctx, path)
	})
}

func DeleteFilesIfExistsInRoot(ctx context.Context, rootPath string, paths []string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.DeleteFilesIfExists(ctx, paths)
	})
}

func DeleteFileAndDirInRoot(ctx context.Context, rootPath string, rootDirs []string, path string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.DeleteFileAndDir(ctx, rootDirs, path)
	})
}

func DeleteFilesAndDirsInRoot(ctx context.Context, rootPath string, rootDirs, paths []string) error {
	return WithRoot(ctx, rootPath, func(lfs FS) error {
		return lfs.DeleteFilesAndDirs(ctx, rootDirs, paths)
	})
}
