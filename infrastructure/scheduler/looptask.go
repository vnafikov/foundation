package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type LoopTask struct {
	Interval       time.Duration
	ExecuteAtStart bool
	Handle         func(ctx context.Context) (logAttrs []slog.Attr, err error)
	Description    string
}
