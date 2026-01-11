package repository

import (
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository struct {
	data  map[string]model.ShortenURL
	mutex sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		data: make(map[string]model.ShortenURL),
	}
}

// TODO: когда появится реальная БД, будет возможность вернуть error, сейчас никогда не вернется, сделал на будущее
func (r *Repository) GetByShortenValue(shortenValue string) *model.ShortenURL {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	shortenUrl, exists := r.data[shortenValue]
	if !exists {
		return nil
	}

	return &shortenUrl
}

func (r *Repository) Add(shortenURL *model.ShortenURL) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.data[shortenURL.ShortenValue] = *shortenURL
}
