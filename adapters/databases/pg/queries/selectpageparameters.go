package queries

import (
	"github.com/vnafikov/common/binarycodec"

	"github.com/vnafikov/foundation/core/parameters/page/pagination"
)

type SelectPageByCursorParameters[T any] struct {
	CursorPagination   pagination.Cursor
	ReadCursor         func(reader binarycodec.Reader) bool
	CursorCapacity     int
	WriteCursor        func(writer *binarycodec.Writer, entity *T) error
	SelectFirstPage    func(limit int) (sql string, args []any)
	SelectNextPage     func(limit int) (sql string, args []any)
	SelectPreviousPage func(limit int) (sql string, args []any)
	SelectLastPage     func(limit int) (sql string, args []any)
	DefaultLimit       int
	MaxLimit           int
}

type SelectPageByOffsetParameters struct {
	OffsetPagination pagination.Offset
	SelectPage       func(offset, limit int) (sql string, args []any)
	DefaultLimit     int
	MaxLimit         int
}
