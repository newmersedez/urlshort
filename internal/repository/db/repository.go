package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

// ErrUniqueViolation возвращается при попытке добавить URL с уже существующим идентификатором.
var ErrUniqueViolation = errors.New("record violates unique constraint")

// DBRepository реализует хранилище сокращённых URL в PostgreSQL.
type DBRepository struct {
	db *DB
}

// NewDBRepository создаёт DBRepository: запускает миграции и устанавливает пул соединений.
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

// Get возвращает сокращённый URL по идентификатору.
// Возвращает nil, nil, если запись не найдена.
func (r *DBRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	row := r.db.pool.QueryRow(
		ctx,
		`SELECT id, user_id, original_value, created_at, is_deleted 
		FROM shorten_urls WHERE id = $1`,
		id)

	var url model.ShortenURL
	err := row.Scan(&url.ID, &url.UserID, &url.OriginalValue, &url.CreatedAt, &url.Deleted)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan query result: %w", err)
	}

	return &url, nil
}

// GetList возвращает все активные (не удалённые) URL указанного пользователя.
func (r *DBRepository) GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	shortenURLs := make([]model.ShortenURL, 0)

	rows, err := r.db.pool.Query(
		ctx,
		`SELECT id, user_id, original_value, created_at, is_deleted 
		FROM shorten_urls WHERE user_id = $1 AND is_deleted = false`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var url model.ShortenURL
		err = rows.Scan(&url.ID, &url.UserID, &url.OriginalValue, &url.CreatedAt, &url.Deleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query result: %w", err)
		}

		shortenURLs = append(shortenURLs, url)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL statement: %w", err)
	}
	return shortenURLs, nil
}

// GetDeletedList возвращает все мягко удалённые URL (используется CleanupService при старте).
func (r *DBRepository) GetDeletedList(ctx context.Context) ([]model.ShortenURL, error) {
	shortenURLs := make([]model.ShortenURL, 0)

	rows, err := r.db.pool.Query(
		ctx,
		`SELECT id, user_id, original_value, created_at, is_deleted 
		FROM shorten_urls WHERE is_deleted = true`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var url model.ShortenURL
		err = rows.Scan(&url.ID, &url.UserID, &url.OriginalValue, &url.CreatedAt, &url.Deleted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query result: %w", err)
		}

		shortenURLs = append(shortenURLs, url)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL statement: %w", err)
	}
	return shortenURLs, nil
}

// Add сохраняет сокращённый URL. Возвращает ErrUniqueViolation, если ID уже существует.
func (r *DBRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	tag, err := r.db.pool.Exec(
		ctx,
		`INSERT INTO shorten_urls (id, user_id, original_value, created_at) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING`,
		shortenURL.ID,
		shortenURL.UserID,
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

// AddBatch сохраняет пакет URL одним INSERT-запросом.
func (r *DBRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	if len(shortenUrls) == 0 {
		return nil
	}

	const argumentsCount = 4

	valueStrings := make([]string, 0, len(shortenUrls))
	valueArgs := make([]any, 0, argumentsCount*len(shortenUrls))

	for i, url := range shortenUrls {
		values := fmt.Sprintf("($%d, $%d, $%d, $%d)", i*argumentsCount+1, i*argumentsCount+2, i*argumentsCount+3, i*argumentsCount+4)
		valueStrings = append(valueStrings, values)

		valueArgs = append(valueArgs, url.ID)
		valueArgs = append(valueArgs, url.UserID)
		valueArgs = append(valueArgs, url.OriginalValue)
		valueArgs = append(valueArgs, url.CreatedAt)
	}

	stmt := fmt.Sprintf("INSERT INTO shorten_urls (id, user_id, original_value, created_at) VALUES %s", strings.Join(valueStrings, ","))
	_, err := r.db.pool.Exec(ctx, stmt, valueArgs...)

	if err != nil {
		return fmt.Errorf("failed to execute SQL stetement: %w", err)
	}

	return nil
}

// SoftDeleteBatch помечает URL пользователя как удалённые (is_deleted = true).
func (r *DBRepository) SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)+1)
	args[0] = userID

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(
		`UPDATE shorten_urls SET is_deleted = true 
		WHERE user_id = $1 AND id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	_, err := r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete batch: %w", err)
	}

	return nil
}

// HardDeleteBatch физически удаляет мягко удалённые URL из базы данных.
func (r *DBRepository) HardDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(
		`DELETE FROM shorten_urls WHERE id IN (%s) AND is_deleted = true`,
		strings.Join(placeholders, ","),
	)

	_, err := r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to hard delete batch: %w", err)
	}

	return nil
}

// Stats возвращает количество сокращённых URL и уникальных пользователей.
func (r *DBRepository) Stats(ctx context.Context) (urls int, users int, err error) {
	row := r.db.pool.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(DISTINCT user_id) FROM shorten_urls WHERE is_deleted = false`)
	if err = row.Scan(&urls, &users); err != nil {
		return 0, 0, fmt.Errorf("failed to get stats: %w", err)
	}
	return urls, users, nil
}

// Ping проверяет доступность базы данных.
func (r *DBRepository) Ping(ctx context.Context) error {
	if err := r.db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to check DB availability: %w", err)
	}

	return nil
}

// Close закрывает пул соединений к базе данных.
func (r *DBRepository) Close() {
	r.db.pool.Close()
}
