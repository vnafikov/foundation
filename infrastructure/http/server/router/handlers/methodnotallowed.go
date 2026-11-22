package handlers

import (
	"net/http"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/clienterrors"
)

func MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	clienterrors.MethodNotAllowed(w)
}
