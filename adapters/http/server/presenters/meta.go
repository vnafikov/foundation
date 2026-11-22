package presenters

type Meta struct {
	CursorPagination cursorPagination `json:"cursorPagination,omitzero"`
}

type cursorPagination struct {
	Next     string `json:"next,omitzero"`
	Previous string `json:"previous,omitzero"`
}

func NewMeta(options MetaOptions) Meta {
	return Meta{CursorPagination: cursorPagination{
		Next:     options.CursorMeta.Next,
		Previous: options.CursorMeta.Previous,
	}}
}
