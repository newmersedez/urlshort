package model

import "time"

type ShortenURL struct {
	ID            string
	OriginalValue string
	CreatedAt     time.Time
}

func NewShortenURL(ID, value string) *ShortenURL {
	return &ShortenURL{
		ID:            ID,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
