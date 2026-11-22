package backgroundtaskparameters

import "github.com/vnafikov/foundation/features/backgroundtasks/core/models"

type SelectAvailableAtByRunner struct {
	Runner        models.BackgroundTaskRunner
	ExcludedIDs   []int64
	LeaseInterval string
}
