package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository struct {
	mu		sync.RWMutex
	file	*os.File
	encoder	*json.Encoder
	data	map[string]model.ShortenURL
}

func NewRepository(filepath string) (*Repository, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	
	if err != nil {
		return nil, err
	}

	repo := &Repository{
		data: make(map[string]model.ShortenURL),
		mu: sync.RWMutex{},
		file: file,
		encoder: json.NewEncoder(file),
	}

	if err := repo.restoreDataFromFile(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Get(ctx context.Context, key string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.data[key]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *Repository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[shortenURL.Key]; !exists {
		r.encoder.Encode(shortenURL)
		r.data[shortenURL.Key] = *shortenURL
	}
	return nil
}

func (r *Repository) Dispose() {
	r.file.Close()
}

func (r *Repository) restoreDataFromFile() error {
	decoder := json.NewDecoder(r.file)
	
	for decoder.More() {
		var url model.ShortenURL
		if err := decoder.Decode(&url); err != nil {
			return nil
		}
		if _, exists := r.data[url.Key]; !exists {
			r.data[url.Key] = url
		}
	}
	
	return nil
}
