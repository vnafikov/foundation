package logger

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/vnafikov/common/filesystem"

	"github.com/vnafikov/foundation/logger/handlers"
	"github.com/vnafikov/foundation/logger/parameters"

	filesystemio "github.com/vnafikov/common/filesystem/io"
)

const (
	rootPath    = "."
	dir         = "logs"
	rootDirPath = rootPath + "/" + dir

	maxFileSize          = 1024 * 1024 * 1024 // 1 GiB.
	maxFiles             = 10
	gzipCompressionLevel = gzip.BestCompression
)

var (
	Level slog.Level

	relPath = filepath.Base(os.Args[0]) + ".log"
	options Options
	f       io.WriteCloser
	l       *slog.Logger
)

func Init(opts Options) error {
	options = opts
	if err := initLogLevel(); err != nil {
		return err
	}

	if options.HasExternalLog {
		if err := initFile(); err != nil {
			return err
		}
	}

	l = slog.New(handler())
	slog.SetDefault(l)
	return nil
}

func initLogLevel() error {
	if options.LogLevel == "" {
		if options.IsDev {
			Level = slog.LevelDebug
			return nil
		}

		Level = slog.LevelInfo
		return nil
	}

	if err := Level.UnmarshalText([]byte(options.LogLevel)); err != nil {
		return fmt.Errorf("%w: %w", ErrSetLogLevel, err)
	}
	return nil
}

func initFile() error {
	if err := filesystem.WithRoot(rootPath, createDir); err != nil {
		return err
	}
	return openFile()
}

func createDir(root *os.Root) error {
	if _, err := root.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := root.Mkdir(dir, filesystem.ModeRWXUserRXGroup); err != nil {
				return fmt.Errorf("%w: %w", ErrCreateDir, err)
			}
			return nil
		}
		return fmt.Errorf("%w: %w", ErrStatDir, err)
	}
	return nil
}

func openFile() error {
	if f != nil {
		return ErrFileAlreadyOpen
	}

	var err error
	f, err = filesystemio.NewRotatingFileWriter(rootDirPath, relPath, maxFileSize, maxFiles, gzipCompressionLevel)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenFile, err)
	}
	return nil
}

func handler() slog.Handler {
	handlerOptions := &slog.HandlerOptions{
		Level:       Level,
		ReplaceAttr: replaceLevel,
	}
	if options.IsDev {
		textHandler := handlers.NewTextHandler(os.Stdout, handlerOptions)
		if !options.HasExternalLog {
			return textHandler
		}

		jsonHandler := slog.NewJSONHandler(f, handlerOptions)
		return slog.NewMultiHandler(textHandler, jsonHandler)
	}
	if !options.HasExternalLog {
		return slog.NewJSONHandler(os.Stdout, handlerOptions)
	}
	return slog.NewJSONHandler(io.MultiWriter(os.Stdout, f), handlerOptions)
}

func replaceLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}

	level, ok := a.Value.Any().(slog.Level)
	if ok && level == parameters.LevelImportant {
		a.Value = slog.StringValue(parameters.LevelImportantString)
	}
	return a
}

func Reopen() error {
	if !options.HasExternalLog {
		return nil
	}

	if err := Close(); err != nil {
		return err
	}
	return Init(options)
}

func Close() error {
	return closeFile()
}

func closeFile() error {
	if f == nil {
		return nil
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrCloseFile, err)
	}

	f = nil
	return nil
}
