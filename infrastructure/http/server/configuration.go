package server

import (
	"time"

	"github.com/julienschmidt/httprouter"
)

type Configuration struct {
	Port              string
	SetRoutes         func(*httprouter.Router)
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int
	ShutdownTimeout   time.Duration
}
