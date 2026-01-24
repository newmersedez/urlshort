package config

import (
	"errors"
	"flag"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
)

var (
	defaultFileStoragePath		= filepath.Join(os.TempDir(), "storage.json")
	errServerAddressNotSet		= errors.New("server address is not set")
	errServerAddressInvalid		= errors.New("server address is not a valid IPv4 address")
	errBaseURLNotSet      		= errors.New("base URL is not set")
	errBaseURLInvalid			= errors.New("base URL is not a valid URL")
	errBaseURLMissingSchema		= errors.New("base URL must have http:// or https:// schema")
	errBaseURLMissingHost		= errors.New("base URL must have valid host")
	errLogLevelNotSet			= errors.New("log level is not set")
	errFileStoragePathNotSet	= errors.New("file storage path is not set")
	errFileStoragePathInvalid	= errors.New("file storage path is not a valid OS path")
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
		return nil, err
	}

	if err := validateServerAddress(cfg.ServerAddr); err != nil {
		return nil, err
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return nil, err
	}
	if err := validateFileStoragePath(cfg.FileStoragePath); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Minimal log level")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.Parse()
}

func parseEnvironment(cfg *Config) error {
	return env.Parse(cfg)
}

func validateServerAddress(serverAddress string) error {
	if serverAddress == "" {
		return errServerAddressNotSet
	}
	if host, port, err := net.SplitHostPort(serverAddress); host == "" || port == "" || err != nil {
		return errServerAddressInvalid
	}
	return nil
}

func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return errBaseURLNotSet
	}
	
	u, err := url.Parse(baseURL)
	if err != nil {
		return errBaseURLInvalid
	}

	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return errBaseURLMissingSchema
	}

	if u.Host == "" {
		return errBaseURLMissingHost
	}

	return nil
}

func validateLogLevel(logLevel string) error {
	if logLevel == "" {
		return errLogLevelNotSet
	}

	return nil
}

func validateFileStoragePath(path string) error {
	if path == "" {
		return errFileStoragePathNotSet
	}

    clean := filepath.Clean(path)
    
    if clean == "." || clean == ".." || strings.HasSuffix(clean, "/") {
        return errFileStoragePathInvalid
    }
    
    // Пробуем создать (но не оставляем) файл по этому пути
    file, err := os.CreateTemp(filepath.Dir(clean), filepath.Base(clean)+"*")
    if err != nil {
        return errFileStoragePathInvalid
    }
    file.Close()
    os.Remove(file.Name())
    
    return nil
}