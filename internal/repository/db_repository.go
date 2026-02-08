package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/newmersedez/urlshort/internal/model"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) (*DBRepository, error) {
	repo := &DBRepository{
		db: db,
	}
	return repo, nil
}

func (r *DBRepository) Get(ctx context.Context, ID string) (*model.ShortenURL, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, original_value, created_at FROM shorten_urls WHERE id = $1", ID)

	var url model.ShortenURL
	err := row.Scan(&url.ID, &url.OriginalValue, &url.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan query result: %w", err)
	}

	return &url, nil
}

func (r *DBRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	result, err := r.db.ExecContext(ctx, "INSERT INTO shorten_urls (id, original_value, created_at) VALUES ($1, $2, $3)", shortenURL.ID, shortenURL.OriginalValue, shortenURL.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	return nil
}

func (r *DBRepository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	return nil
}

func (r *DBRepository) Dispose() {}
