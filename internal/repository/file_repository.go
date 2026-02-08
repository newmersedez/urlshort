package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type FileRepository struct {
	mu      sync.RWMutex
	file    *os.File
	encoder *json.Encoder
	items   map[string]model.ShortenURL
}

func NewFileRepository(filepath string) (*FileRepository, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}

	repo := &FileRepository{
		items:   make(map[string]model.ShortenURL),
		mu:      sync.RWMutex{},
		file:    file,
		encoder: json.NewEncoder(file),
	}

	if err := repo.restoreDataFromFile(); err != nil {
		return nil, fmt.Errorf("failed to restore data from file %s: %w", filepath, err)
	}

	return repo, nil
}

func (r *FileRepository) Get(ctx context.Context, ID string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.items[ID]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *FileRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[shortenURL.ID]; !exists {
		r.encoder.Encode(shortenURL)
		r.items[shortenURL.ID] = *shortenURL
	}
	return nil
}

func (r *FileRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortenURL := range shortenUrls {
		if _, exists := r.items[shortenURL.ID]; !exists {
			r.encoder.Encode(shortenURL)
			r.items[shortenURL.ID] = *shortenURL
		}
	}
	return nil
}

func (r *FileRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *FileRepository) Dispose() {
	r.file.Close()
}

func (r *FileRepository) restoreDataFromFile() error {
	decoder := json.NewDecoder(r.file)

	line := 1
	for decoder.More() {
		var url model.ShortenURL
		if err := decoder.Decode(&url); err != nil {
			return fmt.Errorf("failed to restore url at line %d: %w", line, err)
		}
		if _, exists := r.items[url.ID]; !exists {
			r.items[url.ID] = url
		}
		line++
	}

	return nil
}
