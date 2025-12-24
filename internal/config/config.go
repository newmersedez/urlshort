package config

import (
	"flag"

	handlersConfig "github.com/newmersedez/urlshort/internal/handler/config"
)

type Config struct {
	Handlers handlersConfig.Config
}

func GetConfig() Config {
	cfg := Config{}
	flag.StringVar(&cfg.Handlers.ServerAddr, "addr", "localhost:8080", "address of HTTP server")
	flag.StringVar(&cfg.Handlers.BaseUrl, "baseurl", "http://localhost:8080", "DNS")

	flag.Parse()
	return cfg
}