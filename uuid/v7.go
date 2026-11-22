package uuid

import (
	"encoding/hex"
	"uuid"
)

func V7() string {
	return uuid.NewV7().String()
}

func V7Compact() string {
	id := uuid.NewV7()
	return hex.EncodeToString(id[:])
}
