package backgroundtaskdefinitions

import "github.com/vnafikov/foundation/features/backgroundtasks/core/models"

type Definition struct {
	Name   string
	Runner models.BackgroundTaskRunner
}
