package handlers

import (
	"net/http"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/clienterrors"
)

func NotFound(w http.ResponseWriter, _ *http.Request) {
	clienterrors.NotFound(w)
}
