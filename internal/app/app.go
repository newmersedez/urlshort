package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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

type App struct {
	db         *sql.DB
	cfg        *config.Config
	logger     *slog.Logger
	repository handler.Repository
	shortener  handler.Shortener
}

func NewApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config object: %w", err)
	}

	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger object: %w", err)
	}

	shortener := service.NewURLShortenerService()

	var repo handler.Repository
	var db *sql.DB

	switch {
	case cfg.DatabaseDSN != "":
		db, err = newDatabaseConnection(cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database connection object: %w", err)
		}
		if err := runMigrations(db, logger); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}

		repo, err = repository.NewDBRepository(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize db repository object: %w", err)
		}
	case cfg.FileStoragePath != "":
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file repository object: %w", err)
		}
	default:
		repo, err = repository.NewMemoryRepository()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize in-memory repository object: %w", err)
		}
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		repository: repo,
		shortener:  shortener,
		db:         db,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("Starting server", "address", a.cfg.ServerAddr)
	return handler.Serve(*a.cfg, a.repository, a.shortener, a.logger)
}

func (a *App) Shutdown() error {
	a.repository.Close()

	if a.db != nil {
		return a.db.Close()
	}

	return nil
}

func newDatabaseConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to create DB instance: %w", err)
	}

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(200)

	return db, nil
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
	if db == nil {
		logger.Info("no migrations were applied, skipping...")
		return nil
	}

	logger.Info("Starting migrations...")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize db driver object: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate object: %w", err)
	}

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("finished migrations successfully")
	return nil
}
