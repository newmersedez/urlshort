package main

import (
	"log"

	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	store := repository.NewRepository()
	shortener := service.NewUrlShortenerService()
	logger := log.Default()

	return handler.Serve(cfg.Handlers, store, shortener, logger)

}
