package config

import (
	"errors"
	"flag"
	"net"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	errServerAddressNotSet	= "server address is not set"
	errServerAddressInvalid	= "server address is not a valid IPv4 address"
	errBaseURLNotSet      	= "base URL is not set"
	errBaseURLInvalid		= "base URL is not a valid URL"
	errBaseURLInvalidSchema	= "base URL must have http:// or https:// schema"
	errBaseURLInvalidHost	= "base URL must have valid host"
)

type Config struct {
	ServerAddr string	`env:"SERVER_ADDRESS"`
	BaseURL    string	`env:"BASE_URL"`
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
	
	return &cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "IPv4 address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()
}

func parseEnvironment(cfg *Config) error {
	return env.Parse(cfg)
}

func validateServerAddress(serverAddress string) error {
	if serverAddress == "" {
		return errors.New(errServerAddressNotSet)
	}
	if _, _, err := net.SplitHostPort(serverAddress); err != nil {
		return errors.New(errServerAddressInvalid)
	}
	return nil
}

func validateBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return errors.New(errBaseURLInvalid)
	}

	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return errors.New(errBaseURLInvalidSchema)
	}

	if u.Host == "" {
		return errors.New(errBaseURLInvalidHost)
	}

	return nil
}