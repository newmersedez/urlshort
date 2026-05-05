// Package logger предоставляет вспомогательный конструктор для slog.Logger.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// NewLogger создаёт структурированный JSON-логгер с заданным уровнем логирования.
// Допустимые значения logLevel: "debug", "info", "warn", "error" (без учёта регистра).
// При пустой строке используется уровень Info.
func NewLogger(logLevel string) (*slog.Logger, error) {
	levelVar := new(slog.LevelVar)

	if err := setLogLevel(levelVar, logLevel); err != nil {
		return nil, fmt.Errorf("failed to set log level: %w", err)
	}

	options := &slog.HandlerOptions{
		Level: levelVar,
	}

	handler := slog.NewJSONHandler(os.Stdout, options)
	logger := slog.New(handler)

	return logger, nil
}

func setLogLevel(levelVar *slog.LevelVar, levelStr string) error {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		return fmt.Errorf("failed to determine log level from value %s: %w", levelStr, err)
	}

	levelVar.Set(level)
	return nil
}

func parseLogLevel(levelStr string) (slog.Level, error) {
	levelStr = strings.TrimSpace(strings.ToUpper(levelStr))

	if levelStr == "" {
		return slog.LevelInfo, nil
	}

	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", levelStr)
	}
}
