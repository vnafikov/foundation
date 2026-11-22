package tracer

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type traceCopyFromData struct {
	tableName   pgx.Identifier
	columnNames []string
	startTime   time.Time
}
