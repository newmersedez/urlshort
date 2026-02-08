package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	shortener := service.NewURLShortenerService()

	var repo handler.Repository
	switch {
	case cfg.DatabaseDSN != "":
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to create DB instance: %w", err)
		}
		defer db.Close()
		db.SetMaxOpenConns(200)
		db.SetMaxIdleConns(200)

		repo, err = repository.NewDBRepository(db)
		if err != nil {
			return fmt.Errorf("failed to create repository instance: %w", err)
		}

		logger.Info("Starting migrations...")
		if err = runMigrations(db); err != nil {
			return fmt.Errorf("failed to apply migration: %w", err)
		}
		logger.Info("Finished migrations successfully")
	case cfg.FileStoragePath != "":
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("failed to create repository instance: %w", err)
		}
	default:
		repo, err = repository.NewMemoryRepository()
		if err != nil {
			return fmt.Errorf("failed to create repository instance: %w", err)
		}
	}

	defer repo.Dispose()

	logger.Info("Starting server", "address", cfg.ServerAddr)
	return handler.Serve(*cfg, repo, shortener, logger)
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	// Применяем все миграции
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
