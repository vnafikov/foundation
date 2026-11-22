package queries

type Hydratable[T any] interface {
	Fields() (*T, []any)
}
