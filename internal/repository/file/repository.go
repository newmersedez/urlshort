// Package file реализует хранилище сокращённых URL на основе файла в формате JSONL.
// Данные сохраняются между перезапусками; при старте происходит восстановление из файла.
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

// FileRepository хранит сокращённые URL в памяти и персистирует новые записи в файл.
// Все операции потокобезопасны.
type FileRepository struct {
	mu      sync.RWMutex
	file    *os.File
	encoder *json.Encoder
	items   map[string]model.ShortenURL
}

// NewFileRepository открывает (или создаёт) файл filepath и восстанавливает данные из него.
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

// Get возвращает сокращённый URL по идентификатору.
// Возвращает nil, nil, если запись не найдена.
func (r *FileRepository) Get(ctx context.Context, id string) (*model.ShortenURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortenURL, exists := r.items[id]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

// GetList возвращает все активные (не удалённые) URL указанного пользователя.
func (r *FileRepository) GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
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
func (r *FileRepository) GetDeletedList(ctx context.Context) ([]model.ShortenURL, error) {
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

// Add сохраняет новый URL в память и дописывает его в файл.
// Если ID уже существует, запись пропускается.
func (r *FileRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[shortenURL.ID]; !exists {
		r.encoder.Encode(shortenURL)
		r.items[shortenURL.ID] = *shortenURL
	}
	return nil
}

// AddBatch сохраняет пакет URL за одну операцию блокировки.
// Уже существующие ID пропускаются.
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

// SoftDeleteBatch помечает URL пользователя как удалённые (Deleted = true) в памяти.
// Файл не обновляется - физическое удаление производится через HardDeleteBatch.
func (r *FileRepository) SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error {
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

// HardDeleteBatch физически удаляет мягко удалённые URL из in-memory карты.
func (r *FileRepository) HardDeleteBatch(ctx context.Context, ids []string) error {
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
func (r *FileRepository) Stats(ctx context.Context) (urls int, users int, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uniqueUsers := make(map[uuid.UUID]struct{})
	for _, item := range r.items {
		if item.Deleted {
			continue
		}
		urls++
		uniqueUsers[item.UserID] = struct{}{}
	}
	return urls, len(uniqueUsers), nil
}

// Ping всегда возвращает nil - файловое хранилище всегда доступно.
func (r *FileRepository) Ping(ctx context.Context) error {
	return nil
}

// Close закрывает файл хранилища.
func (r *FileRepository) Close() {
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
