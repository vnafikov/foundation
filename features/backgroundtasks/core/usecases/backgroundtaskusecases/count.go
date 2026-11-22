package backgroundtaskusecases

import (
	"context"

	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/features/backgroundtasks/adapters/queries/backgroundtasks"
)

type Count struct {
	backgroundTaskQueries backgroundtasks.Queries
}

// Handle возвращает количество фоновых задач.
func (useCase *Count) Handle(ctx context.Context, filteringParameters *filtering.Parameters) (int64, error) {
	return useCase.backgroundTaskQueries.SelectCount(ctx, filteringParameters)
}

func NewCount(backgroundTaskQueries backgroundtasks.Queries) Count {
	return Count{backgroundTaskQueries: backgroundTaskQueries}
}
