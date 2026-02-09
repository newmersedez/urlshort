package model

import "time"

type ShortenURL struct {
	Id            string
	OriginalValue string
	CreatedAt     time.Time
}

func NewShortenURL(id, value string) *ShortenURL {
	return &ShortenURL{
		Id:            id,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
