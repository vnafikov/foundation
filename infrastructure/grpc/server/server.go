package server

import (
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/vnafikov/foundation/logger"
)

// Start запускает gRPC-сервер.
func Start(ctx context.Context, cancelCtx context.CancelFunc, configuration Configuration) error {
	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", ":"+configuration.Port)
	if err != nil {
		logger.Error(ctx, err, "net.(*ListenConfig).Listen")

		cancelCtx()
		return err
	}

	srv := grpc.NewServer(configuration.Options...)
	configuration.RegisterServices(srv)
	shutdownErrChan := make(chan error, 1)

	go shutdownOnCancel(ctx, srv, configuration.ShutdownTimeout, shutdownErrChan)

	err = srv.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		err = nil
	}
	if err != nil {
		logger.Error(ctx, err, "grpc.(*Server).Serve")

		cancelCtx()
	}
	return errors.Join(err, <-shutdownErrChan)
}

func shutdownOnCancel(ctx context.Context, srv *grpc.Server, shutdownTimeout time.Duration, errChan chan<- error) {
	<-ctx.Done()
	shutdownCtx, shutdownCancelCtx := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer shutdownCancelCtx()

	stoppedChan := make(chan struct{})

	go gracefulShutdown(srv, stoppedChan)

	select {
	case <-stoppedChan:
		errChan <- nil
	case <-shutdownCtx.Done():
		srv.Stop()
		<-stoppedChan
		err := shutdownCtx.Err()

		logger.Error(shutdownCtx, err, "server.shutdownOnCancel")

		errChan <- err
	}
}

func gracefulShutdown(srv *grpc.Server, stoppedChan chan<- struct{}) {
	srv.GracefulStop()
	close(stoppedChan)
}
