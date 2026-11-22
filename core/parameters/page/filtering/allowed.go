package filtering

type Allowed struct {
	Conditions       AllowedConditions
	CustomConditions AllowedCustomConditions
	Search           AllowedSearch
}
