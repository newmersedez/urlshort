package logger

import "go.uber.org/zap"

var Log *zap.Logger = zap.NewNop();

func Initialize(logLevel string) error {
	level, err := zap.ParseAtomicLevel(logLevel)
	
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.Encoding = "console"
	lg, err := cfg.Build()
	
	if err != nil {
		return err
	}

	Log = lg
	return nil
}
