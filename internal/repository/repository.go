package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository struct {
	db		*sql.DB
	mu		sync.RWMutex
	file	*os.File
	encoder	*json.Encoder
	items	map[string]model.ShortenURL
}

func NewRepository(db *sql.DB, filepath string) (*Repository, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}

	repo := &Repository{
		db: db,
		items: make(map[string]model.ShortenURL),
		mu: sync.RWMutex{},
		file: file,
		encoder: json.NewEncoder(file),
	}

	if err := repo.restoreDataFromFile(); err != nil {
		return nil, fmt.Errorf("failed to restore data from file %s: %w", filepath, err)
	}

	return repo, nil
}

func (r *Repository) Get(ctx context.Context, key string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.items[key]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *Repository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[shortenURL.Key]; !exists {
		r.encoder.Encode(shortenURL)
		r.items[shortenURL.Key] = *shortenURL
	}
	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
    if err := r.db.PingContext(ctx); err != nil {
        return fmt.Errorf("failed to ping DB: %w", err)
    }

	return nil
}

func (r *Repository) Dispose() {
	r.file.Close()
}

func (r *Repository) restoreDataFromFile() error {
	decoder := json.NewDecoder(r.file)
	
	line := 1
	for decoder.More() {
		var url model.ShortenURL
		if err := decoder.Decode(&url); err != nil {
			return fmt.Errorf("failed to restore url at line %d: %w", line, err)
		}
		if _, exists := r.items[url.Key]; !exists {
			r.items[url.Key] = url
		}
		line++
	}
	
	return nil
}
