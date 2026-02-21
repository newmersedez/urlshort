package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
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

func (r *MemoryRepository) GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURLs := make([]model.ShortenURL, 0, len(r.items))
	for _, value := range r.items {
		if value.UserID != userID {
			continue
		}

		shortenURLs = append(shortenURLs, value)
	}

	return shortenURLs, nil
}

func (r *MemoryRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[shortenURL.ID] = *shortenURL
	return nil
}

func (r *MemoryRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortenURL := range shortenUrls {
		r.items[shortenURL.ID] = *shortenURL
	}
	return nil
}

func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) Close() {}
