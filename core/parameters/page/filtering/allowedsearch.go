package filtering

import "slices"

type AllowedSearch []SearchOperator

func (allowedSearch AllowedSearch) IsAllowed(operator SearchOperator) bool {
	return slices.Contains(allowedSearch, operator)
}
