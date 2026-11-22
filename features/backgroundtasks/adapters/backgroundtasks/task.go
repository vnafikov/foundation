package backgroundtasks

import (
	"context"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/definitions/backgroundtaskdefinitions"
)

// TaskHandle выполняет задачу.
// Для отложенного повтора должен возвращать ошибку с методом RetryAt(), не удаляя задачу.
type TaskHandle func(ctx context.Context, id int64, payload []byte) (deleted bool, err error)

/*
Task связывает определение фоновой задачи с обработчиком.

IMPORTANT: Задачи должны быть идемпотентны при повторном выполнении.
*/
type Task struct {
	Definition backgroundtaskdefinitions.Definition
	Handle     TaskHandle
}
