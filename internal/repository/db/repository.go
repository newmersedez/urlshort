package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/newmersedez/urlshort/internal/model"
)

var (
	ErrUniqueViolation = errors.New("record violates unique constraint")
)

type DBRepository struct {
	db *DB
}

func NewDBRepository(ctx context.Context, dsn string, logger *slog.Logger) (*DBRepository, error) {
	db, err := newDB(ctx, dsn, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DB repository object: %w", err)
	}

	repo := &DBRepository{
		db: db,
	}
	return repo, nil
}

func (r *DBRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	row := r.db.pool.QueryRow(
		ctx,
		`SELECT id, original_value, created_at 
		FROM shorten_urls WHERE id = $1`,
		id)

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
	tag, err := r.db.pool.Exec(
		ctx,
		`INSERT INTO shorten_urls (id, original_value, created_at) 
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`,
		shortenURL.ID,
		shortenURL.OriginalValue,
		shortenURL.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	rowsAffectedCount := tag.RowsAffected()
	if rowsAffectedCount != 1 {
		return ErrUniqueViolation
	}

	return nil
}

func (r *DBRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	if len(shortenUrls) == 0 {
		return nil
	}

	const argumentsCount = 3

	valueStrings := make([]string, 0, len(shortenUrls))
	valueArgs := make([]any, 0, argumentsCount*len(shortenUrls))

	for i, url := range shortenUrls {
		values := fmt.Sprintf("($%d, $%d, $%d)", i*argumentsCount+1, i*argumentsCount+2, i*argumentsCount+3)
		valueStrings = append(valueStrings, values)

		valueArgs = append(valueArgs, url.ID)
		valueArgs = append(valueArgs, url.OriginalValue)
		valueArgs = append(valueArgs, url.CreatedAt)
	}

	stmt := fmt.Sprintf("INSERT INTO shorten_urls (id, original_value, created_at) VALUES %s", strings.Join(valueStrings, ","))
	_, err := r.db.pool.Exec(ctx, stmt, valueArgs...)

	if err != nil {
		return fmt.Errorf("failed to execute SQL stetement: %w", err)
	}

	return nil
}

func (r *DBRepository) Ping(ctx context.Context) error {
	if err := r.db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to check DB availability: %w", err)
	}

	return nil
}

func (r *DBRepository) Close() {
	if r.db != nil {
		r.db.pool.Close()
	}
}
