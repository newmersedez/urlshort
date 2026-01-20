package config

import "flag"

type Config struct {
	ServerAddr string
	BaseURL    string
}

func NewConfig() *Config {
	cfg := &Config{}
	
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address of HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened links")
	flag.Parse()

	return cfg
}
