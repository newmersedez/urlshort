package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/newmersedez/urlshort/internal/model"
)

var (
	ErrUniqueViolation = errors.New("record already exists")
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

func (r *DBRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, original_value, created_at 
		FROM shorten_urls WHERE id = $1`,
		id)

	var url model.ShortenURL
	err := row.Scan(&url.Id, &url.OriginalValue, &url.CreatedAt)

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
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`)

	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, shortenURL.Id, shortenURL.OriginalValue, shortenURL.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to insert into table: %w", err)
	}

	if count == 0 {
		return ErrUniqueViolation
	}

	return nil
}

func (r *DBRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	valueStrings := make([]string, 0, len(shortenUrls))
	valueArgs := make([]interface{}, 0, len(shortenUrls)*3)
	for _, url := range shortenUrls {
		valueStrings = append(valueStrings, "($1, $2, $3)")
		valueArgs = append(valueArgs, url.Id)
		valueArgs = append(valueArgs, url.OriginalValue)
		valueArgs = append(valueArgs, url.CreatedAt)
	}

	stmt := fmt.Sprintf("INSERT INTO shorten_urls (id, original_value, created_at) VALUES %s", strings.Join(valueStrings, ","))
	_, err := r.db.ExecContext(ctx, stmt, valueArgs...)

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

func (r *DBRepository) Close() {}
