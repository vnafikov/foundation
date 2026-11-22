package sigint

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vnafikov/foundation/logger"
)

const exitInterruptCode = 2

var interruptSignalChan = make(chan os.Signal, 1)

func Start(ctx context.Context, cancelCtx context.CancelFunc) {
	signal.Notify(interruptSignalChan, os.Interrupt, syscall.SIGTERM)

	go listen(ctx, cancelCtx)
}

func listen(ctx context.Context, cancelCtx context.CancelFunc) {
	select {
	case <-interruptSignalChan: // 1st interrupt signal.
		softExit(ctx, cancelCtx)
	case <-ctx.Done():
	}
	for receivedSignal := range interruptSignalChan { // 2nd interrupt signal.
		if receivedSignal == os.Interrupt {
			hardExit(ctx)
			return
		}
	}
}

func softExit(ctx context.Context, cancelCtx context.CancelFunc) {
	logger.Info(ctx, "The program was interrupted. Gracefully shutting down...")

	cancelCtx()
}

func hardExit(ctx context.Context) {
	logger.Info(ctx, "The program was interrupted again. Exit immediately.")

	os.Exit(exitInterruptCode)
}

func Stop(cancelCtx context.CancelFunc) {
	signal.Stop(interruptSignalChan)
	cancelCtx()
}
