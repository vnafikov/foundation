package handlers

import (
	"context"
	"net/http"

	"github.com/vnafikov/common/stacktrace"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/servererrors"
	"github.com/vnafikov/foundation/logger"
)

const skipStackFramesCount = 6

func Panic(ctx context.Context, w http.ResponseWriter, err any) {
	msg := stacktrace.PanicMessage(err, skipStackFramesCount)
	logger.ErrorMsgWithoutStack(ctx, msg)

	servererrors.InternalServerError(w)
}
