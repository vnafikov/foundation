package backgroundtasks

type Configuration struct {
	PostgresDSN          string
	WorkerID             int
	RunnerConfigurations RunnerConfigurations
	Tasks                []Task
}
