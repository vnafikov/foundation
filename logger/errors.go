package logger

import (
	"errors"
	"fmt"
)

var (
	Err = errors.New("logger error")

	ErrSetLogLevel = fmt.Errorf("%w: cannot set log level", Err)

	ErrCreateDir = fmt.Errorf("%w: cannot create logs directory", Err)
	ErrStatDir   = fmt.Errorf("%w: cannot get stat of logs directory", Err)

	ErrFileAlreadyOpen = fmt.Errorf("%w: log file is already open", Err)
	ErrOpenFile        = fmt.Errorf("%w: cannot open log file", Err)

	ErrCloseFile = fmt.Errorf("%w: cannot close log file", Err)

	ErrWrite = fmt.Errorf("%w: cannot write", Err)
)
