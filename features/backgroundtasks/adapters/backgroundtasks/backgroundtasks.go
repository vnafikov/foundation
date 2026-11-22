package backgroundtasks

import (
	"context"
	"sync"
	"time"

	"github.com/vnafikov/common/retry"

	"github.com/vnafikov/foundation/features/backgroundtasks/adapters/queries/backgroundtasks"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/infrastructure/databases/pg"
)

var infiniteRetryConfiguration = retry.ExponentialConfiguration{
	MinDelay:   2 * time.Second,
	MaxDelay:   time.Hour,
	ResetDelay: 30 * time.Minute,
	MaxJitter:  2 * time.Second,
	MaxRetries: -1,
}

type backgroundTasks struct {
	backgroundTaskQueries backgroundtasks.Queries
	configuration         Configuration
	runnerTaskHandlers    runnerTaskHandlers
	runnerNotifyHandlers  runnerNotifyHandlers
	wg                    sync.WaitGroup
}

func (bt *backgroundTasks) start(ctx context.Context) {
	bt.resetLeases(ctx)
	bt.startRunners(ctx)
	bt.startListener(ctx)
	bt.wg.Wait()
}

func (bt *backgroundTasks) resetLeases(ctx context.Context) {
	_ = retry.Exponential(ctx, infiniteRetryConfiguration, func() error {
		return bt.backgroundTaskQueries.ResetLeases(ctx, bt.configuration.WorkerID)
	})
}

func (bt *backgroundTasks) startRunners(ctx context.Context) {
	for _, runnerName := range models.BackgroundTaskRunners {
		r := newRunner(
			bt.backgroundTaskQueries,
			bt.configuration.WorkerID,
			runnerName,
			bt.configuration.RunnerConfigurations[runnerName],
			bt.runnerTaskHandlers[runnerName],
		)
		bt.runnerNotifyHandlers[r.name] = r.notify

		bt.wg.Add(1)
		go bt.startRunner(ctx, &r)
	}
}

func (bt *backgroundTasks) startRunner(ctx context.Context, r *runner) {
	defer bt.wg.Done()

	r.start(ctx)
}

func (bt *backgroundTasks) startListener(ctx context.Context) {
	l := newListener(bt.configuration.PostgresDSN, bt.runnerNotifyHandlers)
	l.start(ctx)
}

func newBackgroundTasks(backgroundTaskQueries backgroundtasks.Queries, configuration Configuration) backgroundTasks {
	return backgroundTasks{
		backgroundTaskQueries: backgroundTaskQueries,
		configuration:         configuration,
		runnerTaskHandlers:    newRunnerTaskHandlers(configuration.Tasks),
		runnerNotifyHandlers:  make(runnerNotifyHandlers, len(models.BackgroundTaskRunners)),
	}
}

func StartAndWait(ctx context.Context, configuration Configuration) {
	backgroundTaskQueries := backgroundtasks.New(pg.Pool)
	bt := newBackgroundTasks(backgroundTaskQueries, configuration)
	bt.start(ctx)
}
