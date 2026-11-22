package backgroundtasks

import (
	"errors"
	"fmt"
)

var (
	errBackgroundTasks = errors.New("background tasks error")

	errNoBackgroundTaskHandler = fmt.Errorf("%w: no background task handler", errBackgroundTasks)

	errNoRunnerNotifyHandler = fmt.Errorf("%w: no runner notify handler", errBackgroundTasks)
)
