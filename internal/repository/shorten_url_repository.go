package repository

import (
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository struct {
	data map[string]model.ShortenURL
	mu   sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		data: make(map[string]model.ShortenURL),
	}
}

func (r *Repository) Get(key string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.data[key]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *Repository) Add(shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[shortenURL.Key] = *shortenURL
	return nil
}
