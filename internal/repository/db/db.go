package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	pool *sql.DB
}

func newDatabaseConnection(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to create DB instance: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(200)

	return &DB{
		pool: db,
	}, nil
}

func runMigrations(db *sql.DB) error {
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

	return nil
}
