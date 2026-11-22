package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers/request"
)

/*
WithBase оборачивает выполнение базовой предварительной обработки для каждого запроса.

Обработчики выполняются в обратном порядке.
*/
func WithBase(next func(request.BaseHandlerContext)) httprouter.Handle {
	next = withTrace(next)
	next = withRequestID(next)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		next(request.NewBaseHandlerContext(w, r, p))
	}
}
