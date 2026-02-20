package model

import "time"

type ShortenURL struct {
	ID            string
	OriginalValue string
	CreatedAt     time.Time
}

func NewShortenURL(id, value string) *ShortenURL {
	return &ShortenURL{
		ID:            id,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
