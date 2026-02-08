package app

import (
	"database/sql"
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
	cfg        *config.Config
	logger     *slog.Logger
	repository handler.Repository
	shortener  handler.Shortener
}

func NewApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	shortener := service.NewURLShortenerService()

	var repo handler.Repository
	switch {
	case cfg.DatabaseDSN != "":
		repo, err = initDatabaseRepository(cfg.DatabaseDSN, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to init database repository: %w", err)
		}
	case cfg.FileStoragePath != "":
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file repository: %w", err)
		}
	default:
		repo, err = repository.NewMemoryRepository()
		if err != nil {
			return nil, fmt.Errorf("failed to create memory repository: %w", err)
		}
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		repository: repo,
		shortener:  shortener,
	}, nil
}

func (a *App) Run() error {
	defer a.repository.Dispose()

	a.logger.Info("Starting server", "address", a.cfg.ServerAddr)
	return handler.Serve(*a.cfg, a.repository, a.shortener, a.logger)
}

func initDatabaseRepository(dsn string, logger *slog.Logger) (handler.Repository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB instance: %w", err)
	}

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(200)

	repo, err := repository.NewDBRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return repo, nil
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
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

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("Migrations applied successfully")
	return nil
}
