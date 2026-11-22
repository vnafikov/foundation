package sorting

import (
	"slices"

	"github.com/vnafikov/foundation/core/models/meta"
)

type Allowed map[meta.Field][]Direction

func (allowed Allowed) IsAllowed(fieldMeta meta.Field, direction Direction) bool {
	if allowedDirections, ok := allowed[fieldMeta]; ok {
		return slices.Contains(allowedDirections, direction)
	}
	return false
}
