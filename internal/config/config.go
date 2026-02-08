package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	handlersConfig "github.com/newmersedez/urlshort/internal/handler/config"
	loggerConfig "github.com/newmersedez/urlshort/internal/logger/config"
	storageConfig "github.com/newmersedez/urlshort/internal/repository/config"
)

type Config struct {
	Handlers handlersConfig.Config
	Logger   loggerConfig.Config
	Storage  storageConfig.Config
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	parseFlags(&cfg)

	if err := parseEnvironment(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Handlers.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.Handlers.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.Logger.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.Storage.FileStoragePath, "f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.StringVar(&cfg.Storage.DatabaseDSN, "d", "", "Database connection string")
	flag.Parse()
}

func parseEnvironment(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return nil
}
