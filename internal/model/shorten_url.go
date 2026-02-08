package model

import "time"

type ShortenURL struct {
	ID            string
	OriginalValue string
	CreatedAt     time.Time
}

func NewShortenURL(key, value string) *ShortenURL {
	return &ShortenURL{
		ID:            key,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
