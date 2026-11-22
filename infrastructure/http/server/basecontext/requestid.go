package basecontext

import "context"

var ctxKeyRequestID = contextKeyRequestID{}

type contextKeyRequestID struct{}

// WithRequestID записывает ID запроса в контекст.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

/*
RequestID возвращает ID запроса.

Если ID запроса не был записан в контекст, то вернёт пустую строку.
*/
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	id, _ := ctx.Value(ctxKeyRequestID).(string)
	return id
}
