package uploadhelper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vnafikov/common/filesystem"

	"github.com/vnafikov/foundation/adapters/http/server/response"
	"github.com/vnafikov/foundation/features/uploadingfiles/adapters/queries/uploadingfiles"
	"github.com/vnafikov/foundation/features/uploadingfiles/core/models"
	"github.com/vnafikov/foundation/infrastructure/http/server/response/clienterrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/response/servererrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers/request"
	"github.com/vnafikov/foundation/infrastructure/storages/localfs"
	"github.com/vnafikov/foundation/logger"
)

const mimeBufferSize = 512

type UploadFile struct {
	uploadingFileQueries        uploadingfiles.Queries
	supportedMIMETypeExtensions map[string]string
	rootPath                    string
	validate                    func(context.Context, http.ResponseWriter, io.Reader) error
}

/*
Handle создаёт запись о загрузке файла в БД и загружает файл.

Читает заголовок файла.
Проверяет, что файл непустой.
Проверяет MIME тип и определяет расширение файла.
Формирует путь с указанным корнем к файлу с расширением.
Создаёт запись о загрузке файла со сформированным путём к файлу в БД.
Создаёт папки.
Создаёт файл с добавочным временным расширением.
Валидирует и пишет в файл (читает поток и пишет в файл одновременно).
Закрывает файл.
Переименовывает файл, убирает добавочное временное расширение.
Возвращает сформированный путь к файлу.
Если возвращает ошибку после создания файла:
закрывает файл,
удаляет файл (пустые папки останутся),
удаляет запись о загрузке файла из БД.
При отмене контекста прерывает чтение тела запроса.

filePath - путь к файлу без корня и расширения.

IMPORTANT: Обратный прокси ограничивает размер тела запроса и проверяет заголовок Content-Type на [supportedMIMETypeExtensions].
Запись о загрузке файла должна удаляться в транзакции (например, с постановкой задачи).
*/
func (uf *UploadFile) Handle(c request.BaseHandlerContext, filePath string) (path string, err error) {
	cancelReadDoneChan := make(chan struct{})
	stopCancelRead := context.AfterFunc(c.Ctx, func() {
		uf.cancelRead(c.Ctx, c.Writer)
		close(cancelReadDoneChan)
	})
	defer func() {
		if !stopCancelRead() {
			<-cancelReadDoneChan
		}
	}()

	buffer, err := uf.readFileHeader(c.Ctx, c.Writer, c.Request)
	if err != nil {
		return "", err
	}

	if err := uf.validateNonEmpty(c.Ctx, c.Writer, buffer); err != nil {
		return "", err
	}

	extension, err := uf.validateMIMEType(c.Ctx, c.Writer, buffer)
	if err != nil {
		return "", err
	}

	path = uf.rootPath + "/" + filePath + extension
	if err := uf.saveFile(c.Ctx, c.Writer, c.Request, buffer, path); err != nil {
		return "", err
	}
	return path, nil
}

func (*UploadFile) cancelRead(ctx context.Context, w http.ResponseWriter) {
	err := http.NewResponseController(w).SetReadDeadline(time.Now())
	if err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Error(ctx, err, "http.(*ResponseController).SetReadDeadline")
	}
}

func (*UploadFile) readFileHeader(ctx context.Context, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	buffer := make([]byte, mimeBufferSize)
	n, err := io.ReadFull(r.Body, buffer)
	if err != nil && !(errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)) {
		err = fmt.Errorf("%w: %w", ErrReadFileHeader, err)
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, err, "io.ReadFull")
		}
		clienterrors.BadRequest(w)
		return nil, err
	}
	return buffer[:n], nil
}

func (*UploadFile) validateNonEmpty(ctx context.Context, w http.ResponseWriter, buffer []byte) error {
	if len(buffer) == 0 {
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, ErrEmptyFile, "uploadhelper.(*UploadFile).validateNonEmpty")
		}
		clienterrors.UnprocessableContent(w)
		return ErrEmptyFile
	}
	return nil
}

func (uf *UploadFile) validateMIMEType(ctx context.Context, w http.ResponseWriter, buffer []byte) (extension string, err error) {
	mimeType := http.DetectContentType(buffer)
	extension, ok := uf.supportedMIMETypeExtensions[mimeType]
	if !ok {
		if logger.ShouldLogDebug() {
			logger.DebugError(
				ctx,
				ErrUnsupportedMIMEType,
				"uploadhelper.(*UploadFile).validateMIMEType",
				slog.String("mime_type", mimeType),
			)
		}
		clienterrors.UnsupportedMediaType(w)
		return "", ErrUnsupportedMIMEType
	}
	return extension, nil
}

func (uf *UploadFile) saveFile(ctx context.Context, w http.ResponseWriter, r *http.Request, buffer []byte, path string) error {
	err := localfs.WithRoot(ctx, uf.rootPath, func(lfs localfs.FS) error {
		return uf.saveFileInRoot(ctx, w, r, lfs, buffer, path)
	})
	if errors.Is(err, filesystem.ErrWithRoot) {
		servererrors.InternalServerError(w)
	}
	return err
}

func (uf *UploadFile) saveFileInRoot(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	lfs localfs.FS,
	buffer []byte,
	path string,
) (err error) {
	if err := uf.uploadingFileQueries.Insert(ctx, path); err != nil {
		response.WriteError(ctx, w, err)
		return err
	}

	var f *os.File
	ok := false
	tempPath := path + models.UploadingFileTempExtension
	defer func() {
		if f != nil {
			if e := lfs.CloseFile(ctx, f); e != nil {
				err = errors.Join(err, e)
			}
		}
		if !ok {
			if e := uf.cleanup(ctx, lfs, tempPath, path); e != nil {
				err = errors.Join(err, e)
			}
		}
	}()

	dirPath := filepath.Dir(path)
	if err := lfs.CreateDirWithParents(ctx, dirPath); err != nil {
		servererrors.InternalServerError(w)
		return err
	}

	f, err = lfs.CreateFile(ctx, tempPath)
	if err != nil {
		servererrors.InternalServerError(w)
		return err
	}

	if err := uf.validateAndWriteFile(ctx, w, r, buffer, f); err != nil {
		return err
	}

	err = lfs.CloseFile(ctx, f)
	f = nil
	if err != nil {
		servererrors.InternalServerError(w)
		return err
	}

	if err := lfs.RenameFile(ctx, tempPath, path); err != nil {
		servererrors.InternalServerError(w)
		return err
	}

	ok = true
	return nil
}

func (uf *UploadFile) cleanup(ctx context.Context, lfs localfs.FS, tempPath, path string) error {
	if err := lfs.DeleteFileIfExists(ctx, tempPath); err != nil {
		return err
	}
	return uf.uploadingFileQueries.Delete(ctx, path)
}

func (uf *UploadFile) validateAndWriteFile(ctx context.Context, w http.ResponseWriter, r *http.Request, buffer []byte, f *os.File) error {
	fullReader := io.MultiReader(bytes.NewReader(buffer), r.Body)
	teeReader := io.TeeReader(fullReader, f)
	if err := uf.validate(ctx, w, teeReader); err != nil {
		return err
	}

	_, err := io.Copy(f, fullReader)
	if err != nil {
		uf.logCopyError(ctx, err, f.Name())

		servererrors.InternalServerError(w)
	}
	return err
}

func (*UploadFile) logCopyError(ctx context.Context, err error, path string) {
	opError, isOpError := errors.AsType[*net.OpError](err)
	if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) ||
		(isOpError && opError.Op == "read" && opError.Timeout()) {
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, err, "io.Copy", slog.String("path", path))
		}
		return
	}

	logger.Error(ctx, err, "io.Copy", slog.String("path", path))
}

func NewUploadFile(
	uploadingFileQueries uploadingfiles.Queries,
	supportedMIMETypeExtensions map[string]string,
	rootPath string,
	validate func(context.Context, http.ResponseWriter, io.Reader) error,
) UploadFile {
	return UploadFile{
		uploadingFileQueries:        uploadingFileQueries,
		supportedMIMETypeExtensions: supportedMIMETypeExtensions,
		rootPath:                    rootPath,
		validate:                    validate,
	}
}
