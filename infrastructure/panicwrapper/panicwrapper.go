package panicwrapper

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/mitchellh/panicwrap"

	"github.com/vnafikov/foundation/logger"
)

func Start() {
	exitStatus, err := panicwrap.Wrap(&panicwrap.WrapConfig{
		Handler:        handle,
		HidePanic:      true,
		ForwardSignals: []os.Signal{syscall.SIGTERM},
	})
	if err != nil {
		panic(err)
	}

	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}
}

func handle(output string) {
	// Дочерний процесс может удалить лог при ротации, а родительский всё ещё держать его старый дескриптор.
	if err := logger.Reopen(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `%s

%s
`, output, err)
		return
	}

	logger.ErrorMsgWithoutStack(context.Background(), output)
}
