package main

import (
	"fmt"
	"log"

	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to create config instance: %w", err)
	}

	if err := logger.NewLogger(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to create logger instance: %w", err)
	}
	defer logger.Dispose()

	repository, err := repository.NewRepository(cfg.FileStoragePath)
	if err != nil {
		return fmt.Errorf("failed to create repository instance: %w", err)
	}
	defer repository.Dispose()

	shortener := service.NewURLShortenerService()

	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddr))
	return handler.Serve(*cfg, repository, shortener)
}
