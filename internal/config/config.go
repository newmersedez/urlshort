package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddr		string	`env:"SERVER_ADDRESS"`
	BaseURL			string	`env:"BASE_URL"`
	LogLevel		string	`env:"LOG_LEVEL"`
	FileStoragePath	string	`env:"FILE_STORAGE_PATH"`
	DatabaseDSN		string	`env:"DATABASE_DSN"`
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
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.FileStoragePath, "f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "host=localhost user=postgres password=1234 dbname=postgres sslmode=disable", "Database connection string")
	flag.Parse()
}

func parseEnvironment(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return nil
}
