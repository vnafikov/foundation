package backgroundtasks

import "time"

type retryAtError interface {
	error
	RetryAt() time.Time
}
