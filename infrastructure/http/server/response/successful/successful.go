package successful

import (
	"context"
	"net/http"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/render"
)

func OK(ctx context.Context, w http.ResponseWriter, data any) {
	render.WriteJSON(ctx, w, http.StatusOK, data)
}

func Created(w http.ResponseWriter) {
	w.WriteHeader(http.StatusCreated)
}

func CreatedWithBody(ctx context.Context, w http.ResponseWriter, data any) {
	render.WriteJSON(ctx, w, http.StatusCreated, data)
}

func Accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
