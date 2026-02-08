package repository

import (
	"context"
	"database/sql"

	"github.com/newmersedez/urlshort/internal/model"
)

type DBRepository struct {
	db      *sql.DB
}

func NewDBRepository(db *sql.DB) (*DBRepository, error) {
	repo := &DBRepository{
		db:      db,
	}
	return repo, nil
}

func (r *DBRepository) Get(ctx context.Context, key string) (*model.ShortenURL, error) {
}

func (r *DBRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
}

func (r *DBRepository) Ping(ctx context.Context) error {
}

func (r *DBRepository) Dispose() {
}
