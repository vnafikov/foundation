package handlers_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/vnafikov/foundation/logger/handlers"
)

func TestTextHandler(t *testing.T) {
	t.Parallel()

	buffer := new(bytes.Buffer)

	slogtest.Run(t, func(*testing.T) slog.Handler {
		buffer.Reset()
		return handlers.NewTextHandler(buffer, nil)
	}, func(*testing.T) map[string]any {
		return parseLog(t, buffer)
	})
}

func parseLog(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()

	log := buffer.String()
	// Удаляет выделение цветом.
	log = strings.ReplaceAll(log, "\033[0m", "")
	log = strings.ReplaceAll(log, "\033[30m", "")
	log = strings.ReplaceAll(log, "\033[31m", "")
	log = strings.ReplaceAll(log, "\033[32m", "")
	log = strings.ReplaceAll(log, "\033[33m", "")
	log = strings.ReplaceAll(log, "\033[34m", "")
	log = strings.ReplaceAll(log, "\033[35m", "")
	log = strings.ReplaceAll(log, "\033[36m", "")
	log = strings.ReplaceAll(log, "\033[37m", "")
	log = strings.ReplaceAll(log, "\033[90m", "")
	log = strings.ReplaceAll(log, "\033[91m", "")
	log = strings.ReplaceAll(log, "\033[92m", "")
	log = strings.ReplaceAll(log, "\033[93m", "")
	log = strings.ReplaceAll(log, "\033[94m", "")
	log = strings.ReplaceAll(log, "\033[95m", "")
	log = strings.ReplaceAll(log, "\033[96m", "")
	log = strings.ReplaceAll(log, "\033[97m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;27m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;99m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;106m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;202m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;208m", "")
	log = strings.ReplaceAll(log, "\x1b[38;5;214m", "")
	lines := strings.Split(log, "\n")
	result := make(map[string]any)

	// Если первая строка не начинается с пробела и в ней нет ":", то это msg.
	firstLine := lines[0]
	_, hasSpacePrefix := strings.CutPrefix(firstLine, " ")
	if !(hasSpacePrefix || strings.Contains(firstLine, ":")) {
		result[slog.MessageKey] = firstLine
		lines = lines[1:]
	}
	stack := []map[string]any{result}
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Если в строке есть "  ↪", то это метка времени и/или уровень логирования.
		const arrowPrefix = "  ↪"
		if strings.Contains(line, arrowPrefix) {
			s, _ := strings.CutPrefix(line, arrowPrefix)
			if strings.Contains(s, " ") {
				parts := strings.Split(s, " ")
				time := parts[0]
				if time != "" {
					result[slog.TimeKey] = time
				}
				level := parts[1]
				if level != "" {
					result[slog.LevelKey] = level
				}
				continue
			}

			result[slog.TimeKey] = s
			// Заглушка, потому что уровень логирования выводится не всегда.
			result[slog.LevelKey] = slog.LevelInfo.String()
			continue
		}

		// Определяет уровень вложенности по числу пробелов перед ключом.
		depth := 0
		for strings.HasPrefix(line, "  ") {
			depth++
			line = line[2:]
		}
		// Разбирает атрибуты.
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			// Убеждается, что стек соответствует текущей вложенности.
			if len(stack) > depth+1 {
				stack = stack[:depth+1]
			}
			parent := stack[len(stack)-1]
			if _, exists := parent[key]; exists {
				// Если ключ уже существует, то проверяет, является ли он map.
				if subMap, ok := parent[key].(map[string]any); ok {
					stack = append(stack, subMap)
				}
			} else {
				parent[key] = value
			}
			continue
		}

		// Если строка без ": ", то это группа. Создаёт новый уровень вложенности.
		// Убирает ":" из ключа группы.
		key, _ := strings.CutSuffix(line, ":")
		newMap := make(map[string]any)
		// Проверяет, что стек соответствует текущей вложенности.
		if len(stack) > depth+1 {
			stack = stack[:depth+1]
		}
		parent := stack[len(stack)-1]
		parent[key] = newMap
		stack = append(stack, newMap)
	}
	return result
}
