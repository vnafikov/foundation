package bodyhelper

import (
	"encoding/json/v2"
	"log/slog"
	"mime"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/clienterrors"
	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers/request"
	"github.com/vnafikov/foundation/logger"
)

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

func ParseJSONBody(c request.BaseHandlerContext, payload any) error {
	if err := validateContentType(c); err != nil {
		return err
	}

	err := json.UnmarshalRead(c.Request.Body, payload)
	if err != nil {
		if logger.ShouldLogDebug() {
			logger.DebugError(c.Ctx, err, "json.UnmarshalRead")
		}
		clienterrors.BadRequest(c.Writer)
	}
	return err
}

func validateContentType(c request.BaseHandlerContext) error {
	contentType := c.Request.Header.Get(headerContentType)
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		if logger.ShouldLogDebug() {
			logger.DebugError(c.Ctx, err, "mime.ParseMediaType", slog.String("content_type", contentType))
		}
		clienterrors.UnsupportedMediaType(c.Writer)
		return err
	}

	if mediaType != contentTypeJSON {
		if logger.ShouldLogDebug() {
			logger.DebugError(c.Ctx, ErrUnsupportedContentType, "bodyhelper.validateContentType", slog.String("content_type", contentType))
		}
		clienterrors.UnsupportedMediaType(c.Writer)
		return ErrUnsupportedContentType
	}
	return nil
}
