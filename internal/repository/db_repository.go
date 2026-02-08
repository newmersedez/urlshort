package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/newmersedez/urlshort/internal/model"
)

type SQLStatement string

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
	row := r.db.QueryRowContext(ctx,
		`SELECT id, original_value, created_at 
		FROM shorten_urls WHERE id = $1;`,
		ID)

	var url model.ShortenURL
	err := row.Scan(&url.ID, &url.OriginalValue, &url.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan query result: %w", err)
	}

	return &url, nil
}

func (r *DBRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	stmt, err := r.db.PrepareContext(ctx,
		`INSERT INTO shorten_urls (id, original_value, created_at) 
		VALUES ($1, $2, $3);`)
	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortenURL.ID, shortenURL.OriginalValue, shortenURL.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	return nil
}

func (r *DBRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	const batchSize = 1000

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction %w", err)
	}
	defer tx.Rollback()

	stmt, err := r.db.PrepareContext(ctx,
		`INSERT INTO shorten_urls (id, original_value, created_at) 
		VALUES ($1, $2, $3);`)

	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	defer stmt.Close()

	for _, url := range shortenUrls {
		_, err = stmt.ExecContext(ctx, url.ID, url.OriginalValue, url.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert: %w", err)
		}
	}

	return tx.Commit()

}

func (r *DBRepository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	return nil
}

func (r *DBRepository) Dispose() {}
