package pagination

type Cursor struct {
	Next       string
	Previous   string
	IsLastPage bool
	Limit      int
}
