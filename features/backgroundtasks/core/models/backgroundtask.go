//go:generate go tool modelmeta-gen -type=BackgroundTask

package models

import (
	"time"

	"github.com/vnafikov/common/types/nullable"
	"github.com/vnafikov/common/validation"

	"github.com/vnafikov/foundation/core/parameters/page"
	"github.com/vnafikov/foundation/core/parameters/page/filtering"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
)

const (
	BackgroundTaskNameMaxLen             = 100
	BackgroundTaskOrderingKeyMaxLen      = 200
	BackgroundTaskDeduplicationKeyMaxLen = 200

	BackgroundTaskRunnerCompute    BackgroundTaskRunner = "compute"
	BackgroundTaskRunnerFilesystem BackgroundTaskRunner = "filesystem"
	BackgroundTaskRunnerNetwork    BackgroundTaskRunner = "network"
	BackgroundTaskRunnerHeavy      BackgroundTaskRunner = "heavy"
)

var (
	BackgroundTaskRunners = [...]BackgroundTaskRunner{
		BackgroundTaskRunnerCompute,
		BackgroundTaskRunnerFilesystem,
		BackgroundTaskRunnerNetwork,
		BackgroundTaskRunnerHeavy,
	}

	BackgroundTaskDefaultAllowedPageParameters = page.Allowed{
		Filtering: filtering.Allowed{
			Conditions: filtering.AllowedConditions{
				BackgroundTaskMeta.Name: {Operators: filtering.GroupString},
				BackgroundTaskMeta.Runner: {
					Operators: filtering.GroupMembership,
					Validator: filtering.OneOfStrings(BackgroundTaskRunners[:]),
				},
				BackgroundTaskMeta.OrderingKey: {Operators: filtering.GroupString},
				BackgroundTaskMeta.CreatedAt:   {Operators: filtering.GroupRange},
				BackgroundTaskMeta.LeasedByWorkerID: {
					Operators: filtering.GroupNumeric,
					Validator: filtering.NewValidator(validation.IsValidInt32),
				},
				BackgroundTaskMeta.LeasedAt: {Operators: filtering.GroupRange},
			},
		},
	}

	_ queries.Hydratable[BackgroundTask] = BackgroundTask{}
)

type BackgroundTaskRunner string

type BackgroundTask struct {
	ID               int64                `db:"bt.id" api:"id"`
	Name             string               `db:"bt.name" api:"name"`
	Runner           BackgroundTaskRunner `db:"bt.runner" api:"runner"`
	OrderingKey      nullable.String      `db:"bt.ordering_key" api:"orderingKey"`
	DeduplicationKey nullable.String      `db:"bt.deduplication_key" api:"deduplicationKey"`
	Payload          []byte               `db:"bt.payload" api:"payload"`
	CreatedAt        time.Time            `db:"bt.created_at" api:"createdAt"`
	AvailableAt      time.Time            `db:"bt.available_at" api:"availableAt"`
	LeasedByWorkerID nullable.Int         `db:"bt.leased_by_worker_id" api:"leasedByWorkerId"`
	LeasedAt         nullable.Time        `db:"bt.leased_at" api:"leasedAt"`
}
