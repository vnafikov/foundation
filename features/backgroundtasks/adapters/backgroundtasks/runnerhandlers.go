package backgroundtasks

import "github.com/vnafikov/foundation/features/backgroundtasks/core/models"

type runnerNotifyHandlers map[models.BackgroundTaskRunner]func()

type taskHandlers map[string]TaskHandle

type runnerTaskHandlers map[models.BackgroundTaskRunner]taskHandlers

func newRunnerTaskHandlers(tasks []Task) runnerTaskHandlers {
	handlers := make(runnerTaskHandlers, len(models.BackgroundTaskRunners))
	for _, runnerName := range models.BackgroundTaskRunners {
		handlers[runnerName] = make(taskHandlers)
	}
	for _, task := range tasks {
		handlers[task.Definition.Runner][task.Definition.Name] = task.Handle
	}
	return handlers
}
