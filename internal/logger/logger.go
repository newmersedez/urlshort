package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

func NewLogger(logLevel string) error {
	level, err := zap.ParseAtomicLevel(logLevel)

	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.DisableCaller = true

	lg, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %w", err)
	}

	Log = lg
	return nil
}

func Dispose() {
	Log.Sync()
}
