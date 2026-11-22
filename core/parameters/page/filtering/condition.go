package filtering

import (
	"slices"

	"github.com/vnafikov/common/convert"

	"github.com/vnafikov/foundation/core/models/meta"
)

const (
	Equal ConditionOperator = iota
	NotEqual
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	In
	NotIn
	Substring
	Fuzzy
)

var (
	GroupEq     = []ConditionOperator{Equal, NotEqual}
	GroupRange  = []ConditionOperator{GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual}
	GroupIn     = []ConditionOperator{In, NotIn}
	GroupSearch = []ConditionOperator{Substring, Fuzzy}

	GroupMembership            = slices.Concat(GroupEq, GroupIn)
	GroupNumeric               = slices.Concat(GroupEq, GroupRange, GroupIn)
	GroupString                = slices.Concat(GroupEq, GroupIn, GroupSearch)
	GroupAllConditionOperators = slices.Concat(GroupEq, GroupRange, GroupIn, GroupSearch)
)

type ConditionOperator uint8

type Condition struct {
	Field    meta.Field
	Operator ConditionOperator
	Values   []any
}

func NewCondition(
	fieldMeta meta.Field,
	operator ConditionOperator,
	values []any,
	allowed AllowedConditions,
) (condition Condition, ok bool) {
	if len(values) == 0 {
		return Condition{}, false
	}

	allowedCondition, ok := allowed.isAllowed(fieldMeta, operator)
	if !ok {
		return Condition{}, false
	}

	if !slices.Contains(GroupIn, operator) {
		values = values[:1]
	}
	return buildCondition(fieldMeta, operator, values, allowedCondition.Validator)
}

func buildCondition(fieldMeta meta.Field, operator ConditionOperator, values []any, isValid func(any) bool) (condition Condition, ok bool) {
	values = filterValues(values, isValid)
	if len(values) == 0 {
		return Condition{}, false
	}

	condition = Condition{
		Field:    fieldMeta,
		Operator: operator,
		Values:   values,
	}
	return condition, true
}

func NewConditionFromString(
	fieldMeta meta.Field,
	operator string,
	values []string,
	allowed AllowedConditions,
) (condition Condition, ok bool) {
	if operator == "" || len(values) == 0 {
		return Condition{}, false
	}

	condOperator, ok := conditionOperator(operator)
	if !ok {
		return Condition{}, false
	}

	allowedCondition, ok := allowed.isAllowed(fieldMeta, condOperator)
	if !ok {
		return Condition{}, false
	}

	if !slices.Contains(GroupIn, condOperator) {
		values = values[:1]
	}
	parsedValues := convert.ParseBuiltinTypeFromStrings(values, fieldMeta.Type)
	return buildCondition(fieldMeta, condOperator, parsedValues, allowedCondition.Validator)
}

func conditionOperator(operator string) (condOperator ConditionOperator, ok bool) {
	switch operator {
	case "eq":
		return Equal, true
	case "neq":
		return NotEqual, true
	case "gt":
		return GreaterThan, true
	case "gte":
		return GreaterThanOrEqual, true
	case "lt":
		return LessThan, true
	case "lte":
		return LessThanOrEqual, true
	case "in":
		return In, true
	case "nin":
		return NotIn, true
	case "sub":
		return Substring, true
	case "fuzzy":
		return Fuzzy, true
	}
	return Equal, false
}
