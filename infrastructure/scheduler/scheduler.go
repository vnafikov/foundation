package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vnafikov/common/stacktrace"

	"github.com/vnafikov/foundation/logger"
)

const (
	perStartDelay = time.Minute // 1 minute.

	skipStackFramesCount     = 4
	logInfoIntervalThreshold = time.Hour

	loggerAttrKeyTask     = "task"
	loggerAttrKeyDuration = "duration"
)

func StartLoopTasksAndWait(ctx context.Context, tasks []LoopTask) {
	timer := time.NewTimer(0)
	wg := new(sync.WaitGroup)

Loop:
	for _, task := range tasks {
		timer.Reset(perStartDelay)
		select {
		case <-timer.C:
			wg.Add(1)
			go startLoopTask(ctx, wg, task)
		case <-ctx.Done():
			_ = timer.Stop()
			break Loop
		}
	}

	wg.Wait()
}

func startLoopTask(ctx context.Context, wg *sync.WaitGroup, task LoopTask) {
	defer wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	if task.ExecuteAtStart {
		execute(ctx, task)
	}
	for {
		select {
		case <-ticker.C:
			execute(ctx, task)
		case <-ctx.Done():
			return
		}
	}
}

func execute(ctx context.Context, task LoopTask) {
	defer func() {
		if err := recover(); err != nil {
			msg := stacktrace.PanicMessage(err, skipStackFramesCount)
			logger.ErrorMsgWithoutStack(ctx, msg, slog.String(loggerAttrKeyTask, task.Description))
		}
	}()

	isDebug := task.Interval < logInfoIntervalThreshold
	isCompleted := false
	var logAttrs []slog.Attr
	if !isDebug || logger.ShouldLogDebug() {
		logStartTask(ctx, isDebug, task.Description)
		startTime := time.Now()
		defer func() { logTaskCompleted(ctx, isDebug, isCompleted, task.Description, logAttrs, startTime) }()
	}

	logAttrs, err := task.Handle(ctx)
	if err != nil {
		return
	}

	isCompleted = true
}

func logStartTask(ctx context.Context, isDebug bool, description string) {
	msg := "Start scheduled task."
	attr := slog.String(loggerAttrKeyTask, description)
	if isDebug {
		logger.Debug(ctx, msg, attr)
		return
	}

	logger.Info(ctx, msg, attr)
}

func logTaskCompleted(ctx context.Context, isDebug, isCompleted bool, description string, attrs []slog.Attr, startTime time.Time) {
	var msg string
	if isCompleted {
		msg = "Scheduled task completed."
	} else {
		msg = "Scheduled task is incomplete."
	}
	duration := time.Since(startTime)
	newAttrs := make([]slog.Attr, 0, len(attrs)+2) // +2: [loggerAttrKeyTask], [loggerAttrKeyDuration].
	newAttrs = append(newAttrs, slog.String(loggerAttrKeyTask, description))
	newAttrs = append(newAttrs, attrs...)
	newAttrs = append(newAttrs, slog.Duration(loggerAttrKeyDuration, duration))
	if isDebug {
		logger.Debug(ctx, msg, newAttrs...)
		return
	}

	logger.Info(ctx, msg, newAttrs...)
}
