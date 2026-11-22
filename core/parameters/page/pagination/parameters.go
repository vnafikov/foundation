package pagination

const (
	TypeCursor Type = iota
	TypeOffset
)

type Type uint8

type Parameters struct {
	Type   Type
	Cursor Cursor
	Offset Offset
}
