package backgroundtaskdefinitions

import "github.com/vnafikov/foundation/features/backgroundtasks/core/models"

var DeleteFiles = Definition{
	Name:   "delete_files",
	Runner: models.BackgroundTaskRunnerFilesystem,
}

type DeleteFilesPayload struct {
	Paths []string `json:"paths"`
}
