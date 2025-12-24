package repository

import (
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository struct {
	data map[string]model.ShortenUrl
	mutex sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		data: make(map[string]model.ShortenUrl),
	}
}

// TODO: когда появится реальная БД, будет возможность вернуть error, сейчас никогда не вернется, сделал на будущее
func (r *Repository) GetByShortenValue(shortenValue string) (*model.ShortenUrl, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	shortenUrl, exists := r.data[shortenValue]
	if !exists {
		return nil, nil
	}

	return &shortenUrl, nil
}

func (r *Repository) Add(shortenUrl *model.ShortenUrl) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.data[shortenUrl.ShortenValue] = *shortenUrl
	return nil
}