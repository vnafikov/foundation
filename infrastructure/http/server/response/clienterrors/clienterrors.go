package clienterrors

import (
	"context"
	"net/http"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/render"
)

func BadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func Unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}

func Forbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}

func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func MethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func Conflict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
}

func ConflictWithBody(ctx context.Context, w http.ResponseWriter, data any) {
	render.WriteJSON(ctx, w, http.StatusConflict, data)
}

func UnsupportedMediaType(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnsupportedMediaType)
}

func UnprocessableContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnprocessableEntity)
}

func UnprocessableContentWithBody(ctx context.Context, w http.ResponseWriter, data any) {
	render.WriteJSON(ctx, w, http.StatusUnprocessableEntity, data)
}

func TooManyRequests(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
}
