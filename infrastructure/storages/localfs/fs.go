package localfs

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/vnafikov/common/filesystem"

	"github.com/vnafikov/foundation/logger"
)

type FS struct {
	root           *os.Root
	rootPrefix     string
	expectFullPath bool
}

func (lfs FS) CreateDirWithParents(ctx context.Context, dirPath string) error {
	err := lfs.root.MkdirAll(lfs.trimRootPath(dirPath), filesystem.ModeRWXUserRXGroupRXOther)
	if err != nil {
		logger.Error(ctx, err, "os.(*Root).MkdirAll")
	}
	return err
}

func (lfs FS) trimRootPath(path string) string {
	if !lfs.expectFullPath {
		return path
	}
	return strings.TrimPrefix(path, lfs.rootPrefix)
}

func (lfs FS) CreateDirsWithParents(ctx context.Context, dirPaths []string) error {
	dirPaths = filesystem.CompactDirPaths(dirPaths)
	for _, dirPath := range dirPaths {
		if err := lfs.CreateDirWithParents(ctx, dirPath); err != nil {
			return err
		}
	}
	return nil
}

func (lfs FS) CreateFile(ctx context.Context, path string) (*os.File, error) {
	f, err := lfs.root.Create(lfs.trimRootPath(path))
	if err != nil {
		logger.Error(ctx, err, "os.(*Root).Create")

		return nil, err
	}

	if logger.ShouldLogDebug() {
		logger.Debug(ctx, "File was created.", slog.String("path", path))
	}
	return f, nil
}

func (FS) CloseFile(ctx context.Context, f *os.File) error {
	err := f.Close()
	if err != nil {
		logger.Error(ctx, err, "os.(*File).Close")
	}
	return err
}

func (lfs FS) DeleteEmptyDirWithParents(ctx context.Context, rootDirs []string, dirPath string) error {
	dirPath = lfs.trimRootPath(dirPath)
	for {
		dir := filepath.Base(dirPath)
		parent := filepath.Dir(dirPath)
		if slices.Contains(rootDirs, dir) || parent == dirPath {
			return nil
		}

		if err := lfs.root.Remove(dirPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				dirPath = parent
				continue
			}
			if errors.Is(err, syscall.ENOTEMPTY) {
				return nil
			}

			logger.Error(ctx, err, "os.(*Root).Remove")

			return err
		}

		if logger.ShouldLogDebug() {
			logger.Debug(ctx, "Directory was deleted.", slog.String("path", lfs.rootPrefix+dirPath))
		}
		dirPath = parent
	}
}

func (lfs FS) DeleteEmptyDirsWithParents(ctx context.Context, rootDirs, dirPaths []string) error {
	dirPaths = filesystem.CompactDirPaths(dirPaths)
	for _, dirPath := range dirPaths {
		if err := lfs.DeleteEmptyDirWithParents(ctx, rootDirs, dirPath); err != nil {
			return err
		}
	}
	return nil
}

func (lfs FS) DeleteFile(ctx context.Context, path string) error {
	err := lfs.delFile(ctx, path)
	if err != nil {
		logger.Error(ctx, err, "os.(*Root).Remove")
	}
	return err
}

func (lfs FS) delFile(ctx context.Context, path string) error {
	if err := lfs.root.Remove(lfs.trimRootPath(path)); err != nil {
		return err
	}

	if logger.ShouldLogDebug() {
		logger.Debug(ctx, "File was deleted.", slog.String("path", path))
	}
	return nil
}

func (lfs FS) DeleteFileIfExists(ctx context.Context, path string) error {
	err := lfs.delFile(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if logger.ShouldLogDebug() {
				logger.Debug(ctx, "File to delete does not exist.", slog.String("path", path))
			}
			return nil
		}

		logger.Error(ctx, err, "os.(*Root).Remove")
	}
	return err
}

func (lfs FS) DeleteFilesIfExists(ctx context.Context, paths []string) error {
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := lfs.DeleteFileIfExists(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

func (lfs FS) DeleteFileAndDir(ctx context.Context, rootDirs []string, path string) error {
	if err := lfs.DeleteFileIfExists(ctx, path); err != nil {
		return err
	}

	dirPath := filepath.Dir(path)
	return lfs.DeleteEmptyDirWithParents(ctx, rootDirs, dirPath)
}

func (lfs FS) DeleteFilesAndDirs(ctx context.Context, rootDirs, paths []string) error {
	dirPaths := make([]string, len(paths))
	for i, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}

		dirPaths[i] = filepath.Dir(path)
		if err := lfs.DeleteFileIfExists(ctx, path); err != nil {
			return err
		}
	}
	return lfs.DeleteEmptyDirsWithParents(ctx, rootDirs, dirPaths)
}

func (lfs FS) RenameFile(ctx context.Context, oldPath, newPath string) error {
	if err := lfs.root.Rename(lfs.trimRootPath(oldPath), lfs.trimRootPath(newPath)); err != nil {
		logger.Error(ctx, err, "os.(*Root).Rename")

		return err
	}

	if logger.ShouldLogDebug() {
		logger.Debug(ctx, "File was renamed.", slog.String("old_path", oldPath), slog.String("new_path", newPath))
	}
	return nil
}

func New(configuration Configuration) FS {
	lfs := FS{
		root:       configuration.Root,
		rootPrefix: configuration.Root.Name() + "/",
	}
	if configuration.ExpectFullPath == nil {
		lfs.expectFullPath = true
	}
	return lfs
}
