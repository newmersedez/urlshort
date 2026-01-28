package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func NewLogger(logLevel string) (*slog.Logger, error) {
    levelVar := new(slog.LevelVar)
    
    if err := setLogLevel(levelVar, logLevel); err != nil {
        return nil, err
    }

    options := &slog.HandlerOptions{
        Level: levelVar,
    }

	handler := slog.NewTextHandler(os.Stderr, options)
	logger := slog.New(handler)
    
	return logger, nil
}

func setLogLevel(levelVar *slog.LevelVar, levelStr string) error {
    level, err := parseLogLevel(levelStr)
    if err != nil {
        return err
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