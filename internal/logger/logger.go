package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	log	*zap.SugaredLogger
}

func NewLogger(logLevel string) (*Logger, error) {
	level, err := zap.ParseAtomicLevel(logLevel)

	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.DisableCaller = true

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger instance: %w", err)
	}

	return &Logger{log: log.Sugar()}, nil
}

func (l *Logger) Debug(msg string, args...any) {
	l.log.Debugf(msg, args...)
}

func (l *Logger) Info(msg string, args...any) {
	l.log.Infof(msg, args...)
}

func (l *Logger) Warn(msg string, args...any) {
	l.log.Warnf(msg, args...)
}

func (l *Logger) Error(msg string, args...any) {
	l.log.Errorf(msg, args...)
}

func (l *Logger) Dispose() {
	l.log.Sync()
}
