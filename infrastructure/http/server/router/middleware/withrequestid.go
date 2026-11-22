package middleware

import (
	"github.com/vnafikov/foundation/infrastructure/http/server/basecontext"
	"github.com/vnafikov/foundation/infrastructure/http/server/router/handlers/request"
)

const headerAppRequestID = "App-Request-ID"

func withRequestID(next func(request.BaseHandlerContext)) func(request.BaseHandlerContext) {
	return func(c request.BaseHandlerContext) {
		id := c.Request.Header.Get(headerAppRequestID)
		if id != "" {
			c.Ctx = basecontext.WithRequestID(c.Ctx, id)
		}
		next(c)
	}
}
