package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop();

func Initialize(logLevel string) error {
	level, err := zap.ParseAtomicLevel(logLevel)
	
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.DisableCaller = true
	lg, err := cfg.Build()
	
	if err != nil {
		return err
	}

	Log = lg
	return nil
}
