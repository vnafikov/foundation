package tracer

import "time"

type traceQueryData struct {
	sql       string
	args      []any
	startTime time.Time
}
