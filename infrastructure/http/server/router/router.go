package router

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers"
)

func New(ctx context.Context, setRoutes func(*httprouter.Router)) *httprouter.Router {
	r := &httprouter.Router{
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		NotFound:               http.HandlerFunc(handlers.NotFound),
		MethodNotAllowed:       http.HandlerFunc(handlers.MethodNotAllowed),
		PanicHandler:           func(w http.ResponseWriter, _ *http.Request, err any) { handlers.Panic(ctx, w, err) },
	}
	setRoutes(r)
	return r
}
