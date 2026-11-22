package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/vnafikov/foundation/infrastructure/http/server/router"
	"github.com/vnafikov/foundation/logger"
)

/*
Start запускает HTTP-сервер.

IMPORTANT: Обратный прокси ограничивает размер тела запроса.
*/
func Start(ctx context.Context, cancelCtx context.CancelFunc, configuration Configuration) error {
	srv := &http.Server{
		Addr:              ":" + configuration.Port,
		Handler:           router.New(ctx, configuration.SetRoutes),
		ReadTimeout:       configuration.ReadTimeout,
		ReadHeaderTimeout: configuration.ReadHeaderTimeout,
		MaxHeaderBytes:    configuration.MaxHeaderBytes,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}
	shutdownErrChan := make(chan error, 1)

	go shutdownOnCancel(ctx, srv, configuration.ShutdownTimeout, shutdownErrChan)

	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	if err != nil {
		logger.Error(ctx, err, "http.(*Server).ListenAndServe")

		cancelCtx()
	}
	return errors.Join(err, <-shutdownErrChan)
}

func shutdownOnCancel(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration, errChan chan<- error) {
	<-ctx.Done()
	shutdownCtx, shutdownCancelCtx := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer shutdownCancelCtx()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error(shutdownCtx, err, "http.(*Server).Shutdown")

		if e := srv.Close(); e != nil {
			logger.Error(shutdownCtx, e, "http.(*Server).Close")

			err = errors.Join(err, e)
		}
	}
	errChan <- err
}
