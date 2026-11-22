package builders

import (
	"strings"

	"github.com/vnafikov/foundation/core/parameters/page/sorting"
)

func BuildOrderByClause(parameters sorting.Parameters) string {
	parameterCount := len(parameters)
	if parameterCount == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(parameterCount * approxExpressionLen)
	sb.WriteString(`
ORDER BY `)
	for i, parameter := range parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		direction := ""
		if parameter.Direction == sorting.Desc {
			direction = " DESC"
		}
		sb.WriteString(parameter.Field.Column + direction)
	}
	return sb.String()
}
