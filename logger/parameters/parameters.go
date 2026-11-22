package parameters

const (
	DebugErrorMessagePrefix = "Error:"
	HTTPStartMessagePrefix  = "HTTP started:"
	HTTPEndMessagePrefix    = "HTTP completed:"
	GRPCStartMessagePrefix  = "gRPC started:"
	GRPCEndMessagePrefix    = "gRPC completed:"
	PGSQLMessagePrefix      = "PG SQL:"

	StackKey = "stack"

	// LevelImportant позволяет выводить сообщения несмотря на высокий уровень логирования.
	LevelImportant       = 27
	LevelImportantString = "IMPORTANT"
)
