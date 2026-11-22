package backgroundtaskusecases

import (
	"context"

	"github.com/vnafikov/foundation/core/parameters/page"
	"github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/features/backgroundtasks/adapters/queries/backgroundtasks"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
)

type List struct {
	backgroundTaskQueries backgroundtasks.Queries
}

// Handle возвращает фоновые задачи.
func (useCase *List) Handle(ctx context.Context, pageParameters *page.Parameters) ([]*models.BackgroundTask, pagination.CursorMeta, error) {
	return useCase.backgroundTaskQueries.SelectPage(ctx, pageParameters)
}

func NewList(backgroundTaskQueries backgroundtasks.Queries) List {
	return List{backgroundTaskQueries: backgroundTaskQueries}
}
