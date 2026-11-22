package page

import (
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/core/parameters/page/sorting"
)

type Allowed struct {
	Filtering filtering.Allowed
	Sorting   sorting.Allowed
}
