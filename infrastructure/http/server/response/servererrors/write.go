package servererrors

import (
	"context"
	"errors"
	"net/http"
)

func Write(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		ServiceUnavailable(w)
		return
	}

	InternalServerError(w)
}
