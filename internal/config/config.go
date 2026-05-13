// Package config загружает конфигурацию сервиса из флагов командной строки
// и переменных окружения. Переменные окружения имеют приоритет над флагами.
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

// Config хранит параметры запуска сервиса.
// Значения читаются из флагов CLI (-a, -b, -l, -f, -d) и переменных окружения.
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
}

// NewConfig инициализирует Config: сначала парсит флаги CLI, затем переопределяет
// значения переменными окружения. Возвращает ошибку при проблемах с env-парсингом.
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
