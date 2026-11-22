package backgroundtaskparameters

import (
	"github.com/vnafikov/common/types/nullable"

	"github.com/vnafikov/foundation/features/backgroundtasks/core/definitions/backgroundtaskdefinitions"
)

type Insert struct {
	Definition       backgroundtaskdefinitions.Definition
	OrderingKey      nullable.String
	DeduplicationKey nullable.String
	Payload          []byte
	AvailableAt      nullable.Time
}
