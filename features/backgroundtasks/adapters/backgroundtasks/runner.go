package backgroundtasks

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/vnafikov/common/maps"
	"github.com/vnafikov/common/retry"
	"github.com/vnafikov/common/stacktrace"

	"github.com/vnafikov/foundation/features/backgroundtasks/adapters/queries/backgroundtasks"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/queries/backgroundtaskparameters"
	"github.com/vnafikov/foundation/logger"
)

const (
	skipStackFramesCount = 4

	claimErrorRetryDelay = time.Hour
	claimRetryDelay      = 2 * time.Second

	loggerAttrKeyID       = "id"
	loggerAttrKeyName     = "name"
	loggerAttrKeyRunner   = "runner"
	loggerAttrKeyPayload  = "payload"
	loggerAttrKeyDuration = "duration"
)

var retryConfiguration = retry.ExponentialConfiguration{
	MinDelay:   2 * time.Second,
	MaxDelay:   8 * time.Second,
	MaxRetries: 3,
}

type taskGroup []*models.BackgroundTask

type runner struct {
	backgroundTaskQueries backgroundtasks.Queries
	workerID              int
	name                  models.BackgroundTaskRunner
	batchSize             int
	orderedBatchSize      int
	executorCount         int
	leaseInterval         string
	updateLeasesInterval  time.Duration
	taskHandlers          taskHandlers
	taskIDs               maps.RWMutexMap[int64, struct{}]
	hasMore               bool
	threshold             int
	taskGroupChan         chan taskGroup
	executorClaimChan     chan struct{}
	notifyClaimChan       chan struct{}
	claimTimer            *time.Timer
	wg                    sync.WaitGroup
}

func (r *runner) start(ctx context.Context) {
	r.startExecutors(ctx)
	updateLeasesTicker := time.NewTicker(r.updateLeasesInterval)
	defer updateLeasesTicker.Stop()

	r.claimTimer = time.NewTimer(0)
	defer r.claimTimer.Stop()

	r.loop(ctx, updateLeasesTicker)
	r.wg.Wait()
}

func (r *runner) startExecutors(ctx context.Context) {
	for range r.executorCount {
		r.wg.Add(1)
		go r.startExecutor(ctx)
	}
}

func (r *runner) startExecutor(ctx context.Context) {
	defer r.wg.Done()

	for {
		select {
		case group := <-r.taskGroupChan:
			r.executeGroup(ctx, group)
			if group[0].OrderingKey.IsValid {
				r.notify()
			} else {
				r.executorClaim()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *runner) executeGroup(ctx context.Context, group taskGroup) {
	for i, backgroundTask := range group {
		isCompleted := r.execute(ctx, backgroundTask)
		r.taskIDs.Delete(backgroundTask.ID)
		if !isCompleted {
			r.resetGroup(ctx, group[i+1:])
			return
		}
	}
}

func (r *runner) execute(ctx context.Context, backgroundTask *models.BackgroundTask) (isCompleted bool) {
	defer func() {
		if err := recover(); err != nil {
			msg := stacktrace.PanicMessage(err, skipStackFramesCount)
			logger.ErrorMsgWithoutStack(ctx, msg, slog.Any("background_task", backgroundTask))
		}
	}()

	handle, ok := r.taskHandlers[backgroundTask.Name]
	if !ok {
		logger.Error(
			ctx,
			errNoBackgroundTaskHandler,
			"backgroundtasks.(*runner).taskHandlers",
			slog.String(loggerAttrKeyName, backgroundTask.Name),
			slog.String(loggerAttrKeyRunner, string(r.name)),
		)

		return false
	}

	if logger.ShouldLogDebug() {
		r.logStartTask(ctx, backgroundTask)
		startTime := time.Now()
		defer func() { r.logTaskCompleted(ctx, isCompleted, backgroundTask, startTime) }()
	}

	deleted, err := handle(ctx, backgroundTask.ID, backgroundTask.Payload)
	if err != nil {
		if retryAtErr, ok := errors.AsType[retryAtError](err); ok {
			_ = r.backgroundTaskQueries.Reschedule(ctx, &backgroundtaskparameters.Reschedule{
				ID:               backgroundTask.ID,
				AvailableAt:      retryAtErr.RetryAt(),
				LeasedByWorkerID: r.workerID,
			})
		}
		return false
	}

	if !deleted {
		if err := retry.Exponential(ctx, retryConfiguration, func() error {
			return r.backgroundTaskQueries.Delete(ctx, backgroundTask.ID)
		}); err != nil {
			return false
		}
	}
	return true
}

func (r *runner) logStartTask(ctx context.Context, backgroundTask *models.BackgroundTask) {
	logger.Debug(
		ctx,
		"Start background task.",
		slog.Int64(loggerAttrKeyID, backgroundTask.ID),
		slog.String(loggerAttrKeyName, backgroundTask.Name),
		slog.String(loggerAttrKeyRunner, string(r.name)),
		slog.Any(loggerAttrKeyPayload, string(backgroundTask.Payload)),
	)
}

func (r *runner) logTaskCompleted(ctx context.Context, isCompleted bool, backgroundTask *models.BackgroundTask, startTime time.Time) {
	var msg string
	if isCompleted {
		msg = "Background task completed."
	} else {
		msg = "Background task is incomplete."
	}
	duration := time.Since(startTime)
	logger.Debug(
		ctx,
		msg,
		slog.Int64(loggerAttrKeyID, backgroundTask.ID),
		slog.String(loggerAttrKeyName, backgroundTask.Name),
		slog.String(loggerAttrKeyRunner, string(r.name)),
		slog.Duration(loggerAttrKeyDuration, duration),
	)
}

func (r *runner) resetGroup(ctx context.Context, group taskGroup) {
	count := len(group)
	if count == 0 {
		return
	}

	ids := make([]int64, count)
	for i, backgroundTask := range group {
		ids[i] = backgroundTask.ID
	}
	_ = retry.Exponential(ctx, retryConfiguration, func() error {
		return r.backgroundTaskQueries.ResetLeasesByIDs(ctx, &backgroundtaskparameters.ResetLeasesByIDs{
			IDs:              ids,
			LeasedByWorkerID: r.workerID,
		})
	})
	for _, id := range ids {
		r.taskIDs.Delete(id)
	}
}

func (r *runner) notify() {
	select {
	case r.notifyClaimChan <- struct{}{}:
	default:
	}
}

func (r *runner) executorClaim() {
	select {
	case r.executorClaimChan <- struct{}{}:
	default:
	}
}

func (r *runner) loop(ctx context.Context, updateLeasesTicker *time.Ticker) {
	for {
		select {
		case <-r.executorClaimChan:
			if r.hasMore {
				r.claim(ctx)
			} else {
				r.scheduleClaim(ctx)
			}
		case <-r.notifyClaimChan:
			r.claim(ctx)
		case <-r.claimTimer.C:
			r.claim(ctx)
		case <-updateLeasesTicker.C:
			r.updateLeases(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (r *runner) claim(ctx context.Context) {
	r.claimTimer.Stop()
	taskCount := r.taskIDs.Len()
	if taskCount > r.threshold {
		r.hasMore = true
		return
	}

	remainingCapacity := r.batchSize - taskCount
	for remainingCapacity > 0 {
		orderedBatchSize := min(r.orderedBatchSize, remainingCapacity)
		backgroundTaskList, err := r.backgroundTaskQueries.LeaseBatchByRunner(ctx, &backgroundtaskparameters.LeaseBatchByRunner{
			Runner:           r.name,
			ExcludedIDs:      r.taskIDList(),
			LeaseInterval:    r.leaseInterval,
			BatchSize:        remainingCapacity,
			OrderedBatchSize: orderedBatchSize,
			WorkerID:         r.workerID,
		})
		if err != nil {
			r.hasMore = false
			r.claimTimer.Reset(claimErrorRetryDelay)
			return
		}

		if len(backgroundTaskList) == 0 {
			r.hasMore = false
			r.scheduleClaim(ctx)
			return
		}

		for _, backgroundTask := range backgroundTaskList {
			r.taskIDs.Store(backgroundTask.ID, struct{}{})
		}
		if backgroundTaskList[0].OrderingKey.IsValid {
			r.taskGroupChan <- backgroundTaskList
		} else {
			for i := range backgroundTaskList {
				r.taskGroupChan <- backgroundTaskList[i : i+1]
			}
		}
		remainingCapacity -= len(backgroundTaskList)
	}
	r.hasMore = true
}

func (r *runner) scheduleClaim(ctx context.Context) {
	r.claimTimer.Stop()
	availableAt, err := r.backgroundTaskQueries.SelectAvailableAtByRunner(ctx, &backgroundtaskparameters.SelectAvailableAtByRunner{
		Runner:        r.name,
		ExcludedIDs:   r.taskIDList(),
		LeaseInterval: r.leaseInterval,
	})
	if err != nil {
		r.claimTimer.Reset(claimErrorRetryDelay)
		return
	}
	if !availableAt.IsValid {
		return
	}

	// Минимальная задержка предотвращает частые повторы claim, если готовая задача заблокирована другой транзакцией.
	delay := max(time.Until(availableAt.Time), claimRetryDelay)
	r.claimTimer.Reset(delay)
}

func (r *runner) taskIDList() []int64 {
	ids := make([]int64, 0, r.taskIDs.Len())
	r.taskIDs.Range(func(id int64, _ struct{}) bool {
		ids = append(ids, id)
		return true
	})
	return ids
}

func (r *runner) updateLeases(ctx context.Context) {
	_ = retry.Exponential(ctx, infiniteRetryConfiguration, func() error {
		ids := r.taskIDList()
		if len(ids) == 0 {
			return nil
		}
		return r.backgroundTaskQueries.UpdateLeases(ctx, &backgroundtaskparameters.UpdateLeases{
			IDs:              ids,
			LeasedByWorkerID: r.workerID,
		})
	})
}

func newRunner(
	backgroundTaskQueries backgroundtasks.Queries,
	workerID int,
	name models.BackgroundTaskRunner,
	configuration RunnerConfiguration,
	taskHandlers taskHandlers,
) runner {
	return runner{
		backgroundTaskQueries: backgroundTaskQueries,
		workerID:              workerID,
		name:                  name,
		batchSize:             configuration.BatchSize,
		orderedBatchSize:      max(configuration.BatchSize/configuration.ExecutorCount, 1),
		executorCount:         configuration.ExecutorCount,
		leaseInterval:         configuration.LeaseInterval,
		updateLeasesInterval:  configuration.UpdateLeasesInterval,
		taskHandlers:          taskHandlers,
		taskIDs:               maps.NewRWMutexMap[int64, struct{}](configuration.BatchSize),
		threshold:             configuration.BatchSize / 2,
		taskGroupChan:         make(chan taskGroup, configuration.BatchSize),
		executorClaimChan:     make(chan struct{}, 1),
		notifyClaimChan:       make(chan struct{}, 1),
	}
}
