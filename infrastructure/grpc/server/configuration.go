package server

import (
	"time"

	"google.golang.org/grpc"
)

type Configuration struct {
	Port             string
	Options          []grpc.ServerOption
	RegisterServices func(grpc.ServiceRegistrar)
	ShutdownTimeout  time.Duration
}
