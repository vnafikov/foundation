package sorting

import "github.com/vnafikov/foundation/core/models/meta"

const (
	Asc Direction = iota
	Desc
)

var GroupAllDirections = []Direction{Asc, Desc}

type Direction uint8

type Sort struct {
	Field     meta.Field
	Direction Direction
}

func NewSort(fieldMeta meta.Field, direction Direction, allowed Allowed) (sort Sort, ok bool) {
	if !allowed.IsAllowed(fieldMeta, direction) {
		return Sort{}, false
	}

	sort = Sort{
		Field:     fieldMeta,
		Direction: direction,
	}
	return sort, true
}
