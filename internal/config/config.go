package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddr		string	`env:"SERVER_ADDRESS"`
	BaseURL			string	`env:"BASE_URL"`
	LogLevel		string	`env:"LOG_LEVEL"`
	FileStoragePath	string	`env:"FILE_STORAGE_PATH"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	parseFlags(&cfg)
	if err := parseEnvironment(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := validateServerAddress(cfg.ServerAddr); err != nil {
		return nil, fmt.Errorf("failed to validate server address: %w", err)
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("failed to validate base URL: %w", err)
	}
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("failed to validate log level: %w", err)
	}	
	if err := validateFileStoragePath(cfg.FileStoragePath); err != nil {
		return nil, fmt.Errorf("failed to validate file storage path: %w", err)
	}
	
	return &cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.FileStoragePath, "f", filepath.Join(os.TempDir(), "storage.json"), "File storage path")
	flag.Parse()
}

func parseEnvironment(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return nil
}

func validateServerAddress(serverAddress string) error {
	if serverAddress == "" {
		return errors.New("server address is not set")
	}
	if host, port, err := net.SplitHostPort(serverAddress); host == "" || port == "" || err != nil {
		return fmt.Errorf("server address %s is not a valid IPv4 address: %w", serverAddress, err)
	}
	return nil
}

func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return errors.New("base URL is not set")
	}
	
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("base URL %s is not a valid URL: %w", baseURL, err)
	}

	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("base URL %s must have http:// or https:// schema", baseURL)
	}

	if u.Host == "" {
		return fmt.Errorf("base URL %s must have valid host", baseURL)
	}

	return nil
}

func validateLogLevel(logLevel string) error {
	if logLevel == "" {
		return errors.New("log level is not set")
	}

	return nil
}

func validateFileStoragePath(path string) error {
	if path == "" {
		return errors.New("file storage path is not set")
	}

    clean := filepath.Clean(path)
    
    if clean == "." || clean == ".." || strings.HasSuffix(clean, "/") {
        return fmt.Errorf("file storage path %s is not a valid OS path", path)
    }
    
    file, err := os.CreateTemp(filepath.Dir(clean), filepath.Base(clean)+"*")
    if err != nil {
        return fmt.Errorf("file storage path %s is not a valid OS path", path)
    }
    file.Close()
    os.Remove(file.Name())
    
    return nil
}