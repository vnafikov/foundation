package logger

import "log/slog"

// ShouldLogDebug проверяет уровень логирования для Debug.
// Используется для оптимизации, чтобы избежать лишнего создания атрибутов.
func ShouldLogDebug() bool {
	return Level <= slog.LevelDebug
}
