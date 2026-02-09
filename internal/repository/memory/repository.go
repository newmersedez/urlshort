package memory

import (
	"context"
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]model.ShortenURL
}

func NewMemoryRepository() (*MemoryRepository, error) {
	repo := &MemoryRepository{
		items: make(map[string]model.ShortenURL),
		mu:    sync.RWMutex{},
	}
	return repo, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.items[id]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *MemoryRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[shortenURL.Id] = *shortenURL
	return nil
}

func (r *MemoryRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortenURL := range shortenUrls {
		r.items[shortenURL.Id] = *shortenURL
	}
	return nil
}

func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) Close() {}
