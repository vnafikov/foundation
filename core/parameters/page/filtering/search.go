package filtering

import "github.com/vnafikov/common/validation"

const (
	SearchSubstring SearchOperator = iota
	SearchFuzzy
)

type SearchOperator uint8

type Search struct {
	Operator SearchOperator
	Value    string
}

func NewSearch(operator SearchOperator, value string, allowed AllowedSearch) (search Search, ok bool) {
	if !allowed.IsAllowed(operator) || !validation.IsUTF8WithoutNUL(value) {
		return Search{}, false
	}

	search = Search{
		Operator: operator,
		Value:    value,
	}
	return search, true
}
