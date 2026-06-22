// Package config загружает конфигурацию сервиса из нескольких источников.
// Приоритет (от высшего к низшему): env > флаги CLI > файл > дефолты флагов.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/caarlos0/env/v11"
)

// Config хранит параметры запуска сервиса.
// Значения читаются из файла конфигурации (-c/-config / CONFIG),
// флагов CLI (-a, -b, -l, -f, -d, -s) и переменных окружения.
type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"    json:"server_address"`
	BaseURL         string `env:"BASE_URL"           json:"base_url"`
	LogLevel        string `env:"LOG_LEVEL"          json:"log_level"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"  json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN"       json:"database_dsn"`
	AuditFile       string `env:"AUDIT_FILE"         json:"audit_file"`
	AuditURL        string `env:"AUDIT_URL"          json:"audit_url"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"       json:"enable_https"`
	TLSCertCacheDir string `env:"TLS_CERT_CACHE_DIR" json:"tls_cert_cache_dir"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"     json:"trusted_subnet"`
}

func NewConfig() (*Config, error) {
	parseFlags()

	// 1. Переменные окружения
	envCfg := Config{}
	if err := env.Parse(&envCfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// 2. Флаги CLI
	flagCfg := collectCommandLineArgs()

	// 3. Файл конфигурации
	configPath := flag.Lookup("c").Value.String()
	if configPath == "" {
		configPath = flag.Lookup("config").Value.String()
	}

	fileCfg, err := loadFileCfg(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file")
	}

	cfg := Config{}
	flag.VisitAll(func(f *flag.Flag) { applyFlag(&cfg, f) })

	result := Config{}
	for _, src := range []Config{envCfg, flagCfg, fileCfg, cfg} {
		if err := mergo.Merge(&result, src); err != nil {
			return nil, fmt.Errorf("failed to merge config: %w", err)
		}
	}

	return &result, nil
}

func parseFlags() {
	flag.String("c", "", "Path to JSON config file")
	flag.String("config", "", "Path to JSON config file")
	flag.String("a", "localhost:8080", "IPv4 address of HTTP server")
	flag.String("b", "http://localhost:8080", "Base URL for shortened links")
	flag.String("l", "info", "Minimal log level")
	flag.String("f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.String("d", "", "Database connection string")
	flag.String("audit-file", "", "Audit file path")
	flag.String("audit-url", "", "Audit remote server URL")
	flag.Bool("s", false, "Enable HTTPS via Let's Encrypt autocert")
	flag.String("tls-cache-dir", "", "Directory for TLS certificate cache")
	flag.String("t", "", "Trusted subnet in CIDR notation")
	flag.Parse()
}

func collectCommandLineArgs() Config {
	cfg := Config{}
	flag.Visit(func(f *flag.Flag) { applyFlag(&cfg, f) })
	return cfg
}

func applyFlag(cfg *Config, f *flag.Flag) {
	switch f.Name {
	case "a":
		cfg.ServerAddr = f.Value.String()
	case "b":
		cfg.BaseURL = f.Value.String()
	case "l":
		cfg.LogLevel = f.Value.String()
	case "f":
		cfg.FileStoragePath = f.Value.String()
	case "d":
		cfg.DatabaseDSN = f.Value.String()
	case "audit-file":
		cfg.AuditFile = f.Value.String()
	case "audit-url":
		cfg.AuditURL = f.Value.String()
	case "s":
		cfg.EnableHTTPS = f.Value.String() == "true"
	case "tls-cache-dir":
		cfg.TLSCertCacheDir = f.Value.String()
	case "t":
		cfg.TrustedSubnet = f.Value.String()
	}
}

func loadFileCfg(configPath string) (Config, error) {
	if configPath == "" {
		if v, ok := os.LookupEnv("CONFIG"); ok {
			configPath = v
		}
	}
	if configPath == "" {
		return Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}
