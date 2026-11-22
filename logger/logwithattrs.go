package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/vnafikov/foundation/infrastructure/http/server/basecontext"
	"github.com/vnafikov/foundation/logger/parameters"
)

const (
	maxStackFrames     = 10
	skipStackFramesNum = 3

	// minErrorAttrsCap минимальная ёмкость для среза атрибутов с from, request_id, stack.
	minErrorAttrsCap = 3
)

func logWithAttrs(ctx context.Context, level slog.Level, msg string, withStack bool, attrs []slog.Attr) {
	if !l.Enabled(ctx, level) {
		return
	}

	attrs = attrsWithRequestID(ctx, attrs)
	var pc uintptr
	if level == slog.LevelError && withStack {
		pc, attrs = sourcePCAndAttrsWithStack(attrs)
	} else {
		pc = sourcePC()
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	if err := l.Handler().Handle(ctx, r); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", ErrWrite, err)
	}
}

func attrsWithRequestID(ctx context.Context, attrs []slog.Attr) []slog.Attr {
	requestID := basecontext.RequestID(ctx)
	if requestID != "" {
		return append(attrs, slog.String("request_id", requestID))
	}
	return attrs
}

func sourcePCAndAttrsWithStack(attrs []slog.Attr) (uintptr, []slog.Attr) {
	pcs := make([]uintptr, maxStackFrames)
	n := runtime.Callers(skipStackFramesNum, pcs)
	pcs = pcs[:n]
	attrs = append(attrs, slog.Any(parameters.StackKey, stackTrace(pcs)))
	return pcs[0], attrs
}

// stackTrace формирует стек вызовов, исключая фреймы (из пакета runtime) после main.main.
func stackTrace(pcs []uintptr) []slog.Source {
	frames := runtime.CallersFrames(pcs)
	entries := make([]slog.Source, 0, len(pcs))
	for {
		frame, more := frames.Next()
		entries = append(entries, slog.Source{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if frame.Function == "main.main" || !more {
			break
		}
	}
	return entries
}

func sourcePC() uintptr {
	var pcs [1]uintptr
	runtime.Callers(skipStackFramesNum, pcs[:])
	return pcs[0]
}

func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelDebug, msg, true, attrs)
}

func DebugError(ctx context.Context, err error, from string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelDebug, parameters.DebugErrorMessagePrefix+" "+err.Error(), true, errorAttrs(from, attrs))
}

func errorAttrs(from string, attrs []slog.Attr) []slog.Attr {
	newAttrs := make([]slog.Attr, 0, len(attrs)+minErrorAttrsCap)
	newAttrs = append(newAttrs, slog.String("from", from))
	newAttrs = append(newAttrs, attrs...)
	return newAttrs
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelInfo, msg, true, attrs)
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelWarn, msg, true, attrs)
}

func Error(ctx context.Context, err error, from string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelError, err.Error(), true, errorAttrs(from, attrs))
}

func ErrorMsgWithoutStack(ctx context.Context, msg string, attrs ...slog.Attr) {
	logWithAttrs(ctx, slog.LevelError, msg, false, attrs)
}

func Important(ctx context.Context, msg string, attrs ...slog.Attr) {
	logWithAttrs(ctx, parameters.LevelImportant, msg, true, attrs)
}
