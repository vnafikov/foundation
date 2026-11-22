package imagetoavif

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/vnafikov/common/media/ffmpeg/imagetoavif"
	"github.com/vnafikov/common/slices"

	"github.com/vnafikov/foundation/core/media"
	"github.com/vnafikov/foundation/infrastructure/storages/localfs"
	"github.com/vnafikov/foundation/logger"
)

type ImageToAVIF struct {
	it       imagetoavif.ImageToAVIF
	rootPath string
}

func (it *ImageToAVIF) Handle(ctx context.Context, parameters *media.ImageToAVIFParameters) (err error) {
	p := &imagetoavif.Parameters{
		InputPath: parameters.InputPath,
		Outputs:   slices.Transform(parameters.Outputs, it.output),
		Timeout:   parameters.Timeout,
	}
	if err := it.it.Validate(p); err != nil {
		logger.Error(ctx, err, "imagetoavif.(*ImageToAVIF).Validate", slog.String("input_path", parameters.InputPath))

		return err
	}

	ok := false
	defer func() {
		if !ok {
			if e := it.deleteFiles(ctx, parameters.Outputs); e != nil {
				err = errors.Join(err, e)
			}
		}
	}()

	if err := it.createDirs(ctx, parameters.Outputs); err != nil {
		return err
	}

	if err := it.transform(ctx, p); err != nil {
		return err
	}

	ok = true
	return nil
}

func (*ImageToAVIF) output(entity *media.ImageToAVIFOutput) imagetoavif.Output {
	return imagetoavif.Output{
		Width:  entity.Width,
		Height: entity.Height,
		Path:   entity.Path,
	}
}

func (it *ImageToAVIF) deleteFiles(ctx context.Context, outputs []media.ImageToAVIFOutput) error {
	paths := make([]string, len(outputs))
	for i, output := range outputs {
		paths[i] = output.Path
	}
	return localfs.DeleteFilesIfExistsInRoot(ctx, it.rootPath, paths)
}

func (it *ImageToAVIF) createDirs(ctx context.Context, outputs []media.ImageToAVIFOutput) error {
	dirPaths := make([]string, len(outputs))
	for i, output := range outputs {
		dirPaths[i] = filepath.Dir(output.Path)
	}
	return localfs.CreateDirsWithParentsInRoot(ctx, it.rootPath, dirPaths)
}

func (it *ImageToAVIF) transform(ctx context.Context, parameters *imagetoavif.Parameters) error {
	isCompleted := false
	if logger.ShouldLogDebug() {
		it.logStartTransform(ctx, parameters)
		startTime := time.Now()
		defer func() { it.logTransformCompleted(ctx, isCompleted, parameters.InputPath, startTime) }()
	}

	if err := it.it.Transform(ctx, parameters); err != nil {
		logger.Error(ctx, err, "imagetoavif.(*ImageToAVIF).Transform", slog.String("input_path", parameters.InputPath))

		return err
	}

	isCompleted = true
	return nil
}

func (*ImageToAVIF) logStartTransform(ctx context.Context, parameters *imagetoavif.Parameters) {
	logger.Debug(ctx, "Start transforming image to avif.", slog.Any("parameters", parameters))
}

func (*ImageToAVIF) logTransformCompleted(ctx context.Context, isCompleted bool, inputPath string, startTime time.Time) {
	var msg string
	if isCompleted {
		msg = "Transforming image to avif completed."
	} else {
		msg = "Transforming image to avif is incomplete."
	}
	duration := time.Since(startTime)
	logger.Debug(ctx, msg, slog.String("input_path", inputPath), slog.Duration("duration", duration))
}

func New(configuration Configuration) ImageToAVIF {
	return ImageToAVIF{
		it: imagetoavif.New(imagetoavif.Configuration{
			BinPath: configuration.BinPath,
			CRF:     configuration.CRF,
			Preset:  configuration.Preset,
		}),
		rootPath: configuration.RootPath,
	}
}
