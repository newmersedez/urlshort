package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.FileStoragePath, "f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Audit remote server URL")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}
