package tracer

import (
	"errors"
	"fmt"
)

var (
	errTracer = errors.New("PG tracer error")

	errInvalidContextValue = fmt.Errorf("%w: invalid context value", errTracer)

	errContextValueNotFound            = fmt.Errorf("%w: not found", errInvalidContextValue)
	errContextValueTypeAssertionFailed = fmt.Errorf("%w: type assertion failed", errInvalidContextValue)
)
