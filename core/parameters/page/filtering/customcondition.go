package filtering

import "github.com/vnafikov/common/convert"

type CustomConditionName string

type CustomCondition struct {
	Name   CustomConditionName
	Values []any
}

func NewCustomCondition(
	name CustomConditionName,
	values []any,
	allowed AllowedCustomConditions,
) (customCondition CustomCondition, ok bool) {
	allowedCustomCondition, ok := allowed[name]
	if !ok {
		return CustomCondition{}, false
	}

	if allowedCustomCondition.Type == "" {
		if len(values) == 0 {
			return CustomCondition{Name: name}, true
		}
		return CustomCondition{}, false
	}
	return buildCustomCondition(name, values, allowedCustomCondition.Validator)
}

func buildCustomCondition(name CustomConditionName, values []any, isValid func(any) bool) (customCondition CustomCondition, ok bool) {
	values = filterValues(values, isValid)
	if len(values) == 0 {
		return CustomCondition{}, false
	}

	customCondition = CustomCondition{
		Name:   name,
		Values: values,
	}
	return customCondition, true
}

func NewCustomConditionFromString(
	name CustomConditionName,
	values []string,
	allowed AllowedCustomConditions,
) (customCondition CustomCondition, ok bool) {
	if len(values) == 0 {
		return CustomCondition{}, false
	}

	allowedCustomCondition, ok := allowed[name]
	if !ok {
		return CustomCondition{}, false
	}

	if allowedCustomCondition.Type == "" {
		if len(values) == 1 && values[0] == "" {
			return CustomCondition{Name: name}, true
		}
		return CustomCondition{}, false
	}

	parsedValues := convert.ParseBuiltinTypeFromStrings(values, allowedCustomCondition.Type)
	return buildCustomCondition(name, parsedValues, allowedCustomCondition.Validator)
}
