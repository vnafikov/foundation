package logger

type Options struct {
	IsDev          bool
	HasExternalLog bool
	LogLevel       string // See [slog.Level].
}
