package filtering

type AllowedCustomConditions map[CustomConditionName]AllowedCustomCondition

type AllowedCustomCondition struct {
	Type      string
	Validator func(any) bool
}
