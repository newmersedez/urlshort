package main

import (
	"fmt"
	"log"

	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config object: %w", err)
	}

	log, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger object: %w", err)
	}

	shortener := service.NewURLShortenerService()

	var repo handler.Repository

	switch {
	case cfg.DatabaseDSN != "":
		repo, err = repository.NewDBRepository(cfg.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to initialize db repository object: %w", err)
		}
	case cfg.FileStoragePath != "":
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("failed to initialize file repository object: %w", err)
		}
	default:
		repo, err = repository.NewMemoryRepository()
		if err != nil {
			return fmt.Errorf("failed to initialize in-memory repository object: %w", err)
		}
	}

	defer repo.Close()

	log.Info("Starting server", "address", cfg.ServerAddr)
	return handler.Serve(*cfg, repo, shortener, log)
}
