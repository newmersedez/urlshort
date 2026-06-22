// Package memory реализует хранилище сокращённых URL на основе синхронизированной карты в памяти.
// Данные не сохраняются при перезапуске. Подходит для разработки и тестирования.
package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

// MemoryRepository хранит сокращённые URL в памяти.
// Все операции потокобезопасны благодаря RWMutex.
type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]model.ShortenURL
}

// NewMemoryRepository создаёт новый пустой MemoryRepository.
func NewMemoryRepository() (*MemoryRepository, error) {
	repo := &MemoryRepository{
		items: make(map[string]model.ShortenURL),
		mu:    sync.RWMutex{},
	}
	return repo, nil
}

// Get возвращает сокращённый URL по идентификатору.
// Возвращает nil, nil, если запись не найдена.
func (r *MemoryRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.items[id]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

// GetList возвращает все активные (не удалённые) URL указанного пользователя.
func (r *MemoryRepository) GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURLs := make([]model.ShortenURL, 0, len(r.items))
	for _, value := range r.items {
		if value.UserID != userID || value.Deleted {
			continue
		}

		shortenURLs = append(shortenURLs, value)
	}

	return shortenURLs, nil
}

// GetDeletedList возвращает все мягко удалённые URL (используется CleanupService при старте).
func (r *MemoryRepository) GetDeletedList(ctx context.Context) ([]model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURLs := make([]model.ShortenURL, 0, len(r.items))
	for _, value := range r.items {
		if value.Deleted {
			shortenURLs = append(shortenURLs, value)
		}
	}

	return shortenURLs, nil
}

// Add сохраняет новый сокращённый URL. Если ID уже существует - перезаписывает запись.
func (r *MemoryRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[shortenURL.ID] = *shortenURL
	return nil
}

// AddBatch сохраняет пакет сокращённых URL за одну операцию блокировки.
func (r *MemoryRepository) AddBatch(ctx context.Context, shortenUrls []*model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortenURL := range shortenUrls {
		r.items[shortenURL.ID] = *shortenURL
	}
	return nil
}

// SoftDeleteBatch помечает URL пользователя как удалённые (Deleted = true).
// URL с чужим UserID пропускаются.
func (r *MemoryRepository) SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range ids {
		if url, exists := r.items[id]; exists && url.UserID == userID {
			url.Deleted = true
			r.items[id] = url
		}
	}
	return nil
}

// HardDeleteBatch физически удаляет мягко удалённые URL из хранилища.
// Записи, у которых Deleted = false, не затрагиваются.
func (r *MemoryRepository) HardDeleteBatch(ctx context.Context, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range ids {
		if url, exists := r.items[id]; exists && url.Deleted {
			delete(r.items, id)
		}
	}
	return nil
}

// Stats возвращает количество сокращённых URL и уникальных пользователей.
func (r *MemoryRepository) Stats(ctx context.Context) (urls int, users int, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uniqueUsers := make(map[uuid.UUID]struct{})
	for _, item := range r.items {
		uniqueUsers[item.UserID] = struct{}{}
	}
	return len(r.items), len(uniqueUsers), nil
}

// Ping всегда возвращает nil - in-memory хранилище всегда доступно.
func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

// Close - заглушка; in-memory хранилище не требует освобождения ресурсов.
func (r *MemoryRepository) Close() {}
