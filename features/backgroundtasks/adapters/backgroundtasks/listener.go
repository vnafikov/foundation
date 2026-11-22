package backgroundtasks

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vnafikov/common/retry"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/infrastructure/databases/pg"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
	"github.com/vnafikov/foundation/logger"
)

const channel = "background_tasks"

type listener struct {
	dsn                  string
	runnerNotifyHandlers runnerNotifyHandlers
}

func (l *listener) start(ctx context.Context) {
	_ = retry.Exponential(ctx, infiniteRetryConfiguration, func() error {
		return l.handle(ctx)
	})
}

func (l *listener) handle(ctx context.Context) (err error) {
	conn, err := pg.Connect(ctx, l.dsn)
	if err != nil {
		return err
	}
	defer func() {
		if e := pg.CloseConn(ctx, conn); e != nil {
			err = errors.Join(err, e)
		}
	}()

	q := queries.NewConnQueries(conn)
	if err := q.Listen(ctx, channel); err != nil {
		return err
	}

	l.notifyAllRunners()
	for {
		notification, err := q.WaitForNotification(ctx)
		if err != nil {
			return err
		}

		l.notifyRunner(ctx, notification.Payload)
	}
}

func (l *listener) notifyAllRunners() {
	for _, notify := range l.runnerNotifyHandlers {
		notify()
	}
}

func (l *listener) notifyRunner(ctx context.Context, name string) {
	notify, ok := l.runnerNotifyHandlers[models.BackgroundTaskRunner(name)]
	if !ok {
		logger.Error(
			ctx,
			errNoRunnerNotifyHandler,
			"backgroundtasks.(*listener).runnerNotifyHandlers",
			slog.String(loggerAttrKeyRunner, name),
		)

		return
	}

	notify()
}

func newListener(dsn string, runnerNotifyHandlers runnerNotifyHandlers) listener {
	return listener{
		dsn:                  dsn,
		runnerNotifyHandlers: runnerNotifyHandlers,
	}
}
