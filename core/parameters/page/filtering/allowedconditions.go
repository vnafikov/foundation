package filtering

import (
	"slices"

	"github.com/vnafikov/foundation/core/models/meta"
)

type AllowedCondition struct {
	Operators []ConditionOperator
	Validator func(any) bool
}

type AllowedConditions map[meta.Field]AllowedCondition

func (allowedConditions AllowedConditions) isAllowed(
	fieldMeta meta.Field,
	operator ConditionOperator,
) (allowedCondition AllowedCondition, ok bool) {
	if allowedCondition, ok := allowedConditions[fieldMeta]; ok {
		return allowedCondition, slices.Contains(allowedCondition.Operators, operator)
	}
	return AllowedCondition{}, false
}
