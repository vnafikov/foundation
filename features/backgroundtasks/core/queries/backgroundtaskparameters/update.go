package backgroundtaskparameters

import (
	"time"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
)

type LeaseBatchByRunner struct {
	Runner           models.BackgroundTaskRunner
	ExcludedIDs      []int64
	LeaseInterval    string
	BatchSize        int
	OrderedBatchSize int
	WorkerID         int
}

type UpdateLeases struct {
	IDs              []int64
	LeasedByWorkerID int
}

type ResetLeasesByIDs struct {
	IDs              []int64
	LeasedByWorkerID int
}

type Reschedule struct {
	ID               int64
	AvailableAt      time.Time
	LeasedByWorkerID int
}
