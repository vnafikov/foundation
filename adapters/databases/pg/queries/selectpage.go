package queries

import (
	"context"

	"github.com/vnafikov/common/binarycodec"

	paginationparameters "github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries/pagination"
	"github.com/vnafikov/foundation/logger"
)

/*
selectPageByCursor возвращает срез записей по курсору и курсоры.

NOTE: Подразумевается, что курсор игнорирует пользовательскую сортировку.
*/
func selectPageByCursor[T any](
	ctx context.Context,
	parameters *SelectPageByCursorParameters[T],
	queryRows func(ctx context.Context, sql string, args []any) ([]*T, error),
) ([]*T, paginationparameters.CursorMeta, error) {
	isLastPage := parameters.CursorPagination.IsLastPage
	isNextPage := false
	isPreviousPage := false
	if !isLastPage {
		switch {
		case parameters.CursorPagination.Next != "":
			cursor, err := binarycodec.NewReaderFromEncoded(parameters.CursorPagination.Next)
			if err != nil {
				break
			}

			isNextPage = parameters.ReadCursor(cursor)
		case parameters.CursorPagination.Previous != "":
			cursor, err := binarycodec.NewReaderFromEncoded(parameters.CursorPagination.Previous)
			if err != nil {
				break
			}

			isPreviousPage = parameters.ReadCursor(cursor)
		}
	}
	var selectPage func(int) (string, []any)
	switch {
	case isLastPage:
		selectPage = parameters.SelectLastPage
	case isNextPage:
		selectPage = parameters.SelectNextPage
	case isPreviousPage:
		selectPage = parameters.SelectPreviousPage
	default:
		selectPage = parameters.SelectFirstPage
	}
	limit := pagination.NormalizeLimit(pagination.NormalizeLimitOptions{
		Limit:        parameters.CursorPagination.Limit,
		DefaultLimit: parameters.DefaultLimit,
		MaxLimit:     parameters.MaxLimit,
	})
	sql, args := selectPage(limit)
	entities, err := queryRows(ctx, sql, args)
	if err != nil {
		return nil, paginationparameters.CursorMeta{}, err
	}

	entityCount := len(entities)
	if entityCount == 0 {
		return entities, paginationparameters.CursorMeta{}, nil
	}

	var cursorMeta paginationparameters.CursorMeta
	switch {
	case isNextPage:
		cursorMeta.Previous = encodeCursor(ctx, parameters.CursorCapacity, entities[0], parameters.WriteCursor)
	case isPreviousPage:
		cursorMeta.Next = encodeCursor(ctx, parameters.CursorCapacity, entities[entityCount-1], parameters.WriteCursor)
	}
	if entityCount > limit {
		if isLastPage || isPreviousPage {
			cursorMeta.Previous = encodeCursor(ctx, parameters.CursorCapacity, entities[1], parameters.WriteCursor)
			entities = entities[1 : limit+1]
		} else {
			cursorMeta.Next = encodeCursor(ctx, parameters.CursorCapacity, entities[limit-1], parameters.WriteCursor)
			entities = entities[:limit]
		}
	}
	return entities, cursorMeta, nil
}

func encodeCursor[T any](ctx context.Context, capacity int, entity *T, write func(*binarycodec.Writer, *T) error) string {
	writer := binarycodec.NewWriter(capacity)
	if err := write(&writer, entity); err != nil {
		logger.Error(ctx, err, "queries.encodeCursor.write")

		return ""
	}
	return writer.Encode()
}

// selectPageByOffset возвращает срез записей по смещению.
func selectPageByOffset[T any](
	ctx context.Context,
	parameters *SelectPageByOffsetParameters,
	queryRows func(ctx context.Context, sql string, args []any) ([]*T, error),
) ([]*T, error) {
	pageIndex := max(parameters.OffsetPagination.Page-1, 0)
	limit := pagination.NormalizeLimit(pagination.NormalizeLimitOptions{
		Limit:        parameters.OffsetPagination.Limit,
		DefaultLimit: parameters.DefaultLimit,
		MaxLimit:     parameters.MaxLimit,
	})
	offset := pageIndex * limit
	sql, args := parameters.SelectPage(offset, limit)
	return queryRows(ctx, sql, args)
}
