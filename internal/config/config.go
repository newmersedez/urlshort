// Package config загружает конфигурацию сервиса из файла JSON, флагов командной строки
// и переменных окружения. Приоритет: переменные окружения > флаги CLI > файл конфигурации.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

// Config хранит параметры запуска сервиса.
// Значения читаются из файла конфигурации (-c/-config / CONFIG),
// флагов CLI (-a, -b, -l, -f, -d, -s) и переменных окружения.
type Config struct {
	// ServerAddr - адрес и порт HTTP-сервера (например, "localhost:8080").
	ServerAddr string `env:"SERVER_ADDRESS"`
	// BaseURL - базовый URL для формирования коротких ссылок (например, "http://localhost:8080").
	BaseURL string `env:"BASE_URL"`
	// LogLevel - минимальный уровень логирования (debug, info, warn, error).
	LogLevel string `env:"LOG_LEVEL"`
	// FileStoragePath - путь к файлу JSON для хранения ссылок (используется, если не задан DatabaseDSN).
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// DatabaseDSN - строка подключения к PostgreSQL. Если задана, файловое хранилище не используется.
	DatabaseDSN string `env:"DATABASE_DSN"`
	// AuditFile - путь к файлу аудит-лога. Если пуст - FileAuditObserver не подключается.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL - URL удалённого сервера аудит-событий. Если пуст - HTTPAuditObserver не подключается.
	AuditURL string `env:"AUDIT_URL"`
	// EnableHTTPS - если true, сервер запускается с TLS через autocert (Let's Encrypt).
	EnableHTTPS bool `env:"ENABLE_HTTPS"`
	// TLSCertCacheDir - директория для кэша TLS-сертификатов Let's Encrypt.
	TLSCertCacheDir string `env:"TLS_CERT_CACHE_DIR"`
}

type fileConfig struct {
	ServerAddr      string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	LogLevel        string `json:"log_level"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	EnableHTTPS     bool   `json:"enable_https"`
	TLSCertCacheDir string `json:"tls_cert_cache_dir"`
}

// NewConfig инициализирует Config: сначала парсит флаги CLI, затем переопределяет
// значения переменными окружения. Возвращает ошибку при проблемах с env-парсингом.
func NewConfig() (*Config, error) {
	cfg := Config{}

	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to JSON config file")
	flag.StringVar(&configPath, "config", "", "Path to JSON config file")
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.FileStoragePath, "f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Audit remote server URL")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "Enable HTTPS via Let's Encrypt autocert")
	flag.StringVar(&cfg.TLSCertCacheDir, "tls-cache-dir", "", "Directory for TLS certificate cache")
	flag.Parse()

	if configPath == "" {
		if v, ok := os.LookupEnv("CONFIG"); ok {
			configPath = v
		}
	}

	if configPath != "" {
		fc, err := loadFileConfig(configPath)
		if err != nil {
			return nil, err
		}
		applyFileConfig(&cfg, fc)
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}

func loadFileConfig(path string) (*fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fc fileConfig
	if err = json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &fc, nil
}

func applyFileConfig(cfg *Config, fc *fileConfig) {
	set := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })

	if !set["a"] && fc.ServerAddr != "" {
		cfg.ServerAddr = fc.ServerAddr
	}
	if !set["b"] && fc.BaseURL != "" {
		cfg.BaseURL = fc.BaseURL
	}
	if !set["l"] && fc.LogLevel != "" {
		cfg.LogLevel = fc.LogLevel
	}
	if !set["f"] && fc.FileStoragePath != "" {
		cfg.FileStoragePath = fc.FileStoragePath
	}
	if !set["d"] && fc.DatabaseDSN != "" {
		cfg.DatabaseDSN = fc.DatabaseDSN
	}
	if !set["audit-file"] && fc.AuditFile != "" {
		cfg.AuditFile = fc.AuditFile
	}
	if !set["audit-url"] && fc.AuditURL != "" {
		cfg.AuditURL = fc.AuditURL
	}
	if !set["s"] && fc.EnableHTTPS {
		cfg.EnableHTTPS = fc.EnableHTTPS
	}
	if !set["tls-cache-dir"] && fc.TLSCertCacheDir != "" {
		cfg.TLSCertCacheDir = fc.TLSCertCacheDir
	}
}
