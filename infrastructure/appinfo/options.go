package appinfo

import "github.com/vnafikov/common/types/nullable"

type Options struct {
	IsDev            nullable.Bool
	HasExternalLog   nullable.Bool
	Version          nullable.String
	HTTPPort         nullable.String
	GRPCPort         nullable.String
	WorkerID         nullable.Int
	DisableScheduler nullable.Bool
	ScheduledTasks   nullable.Int
	BaseDomain       nullable.String
	Extras           map[string]string
}
