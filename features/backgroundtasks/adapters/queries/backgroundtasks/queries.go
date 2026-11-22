package backgroundtasks

import (
	adapterqueries "github.com/vnafikov/foundation/adapters/databases/pg/queries"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/models"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
)

type Queries struct {
	q adapterqueries.Queries[models.BackgroundTask]
}

func New(conn queries.Querier) Queries {
	var q Queries
	q.q.UseConn(conn)
	return q
}
