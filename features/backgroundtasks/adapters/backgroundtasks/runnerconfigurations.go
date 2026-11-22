package backgroundtasks

import (
	"time"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
)

type RunnerConfiguration struct {
	BatchSize            int
	ExecutorCount        int
	LeaseInterval        string
	UpdateLeasesInterval time.Duration
}

type RunnerConfigurations map[models.BackgroundTaskRunner]RunnerConfiguration
