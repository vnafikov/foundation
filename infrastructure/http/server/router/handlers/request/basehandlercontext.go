package request

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type BaseHandlerContext struct {
	Ctx            context.Context
	Writer         *ResponseWriter
	Request        *http.Request
	PathParameters httprouter.Params
}

func NewBaseHandlerContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) BaseHandlerContext {
	return BaseHandlerContext{
		Ctx:            r.Context(),
		Writer:         &ResponseWriter{ResponseWriter: w},
		Request:        r,
		PathParameters: p,
	}
}
