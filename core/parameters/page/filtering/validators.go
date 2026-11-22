package filtering

import (
	"slices"

	"github.com/vnafikov/common/validation"
)

func filterValues(values []any, isValid func(any) bool) []any {
	return slices.DeleteFunc(values, func(value any) bool {
		s, ok := value.(string)
		if ok && !validation.IsUTF8WithoutNUL(s) {
			return true
		}
		return isValid != nil && value != nil && !isValid(value)
	})
}

func NewValidator[T any](isValid func(T) bool) func(any) bool {
	return func(value any) bool {
		v, ok := value.(T)
		return ok && isValid(v)
	}
}

func OneOfStrings[T ~string](allowedValues []T) func(any) bool {
	return NewValidator(func(value string) bool {
		return slices.Contains(allowedValues, T(value))
	})
}
