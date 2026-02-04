package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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
		return fmt.Errorf("failed to create config instance: %w", err)
	}

	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create logger instance: %w", err)
	}

    db, err := sql.Open("pgx", cfg.DatabaseDSN)
    if err != nil {
		return fmt.Errorf("failed to create DB instance: %w", err)
    }
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second);
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	repository, err := repository.NewRepository(db, cfg.FileStoragePath)
	if err != nil {
		return fmt.Errorf("failed to create repository instance: %w", err)
	}
	defer repository.Dispose()

	shortener := service.NewURLShortenerService()

	logger.Info("Starting server", "address", cfg.ServerAddr)
	return handler.Serve(*cfg, repository, shortener, logger)
}
