package page

import (
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/core/parameters/page/pagination"
	"github.com/vnafikov/foundation/core/parameters/page/sorting"
)

type Parameters struct {
	Pagination pagination.Parameters
	Filtering  filtering.Parameters
	Sorting    sorting.Parameters
}
