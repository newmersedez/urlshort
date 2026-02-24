package model

import (
	"time"

	"github.com/google/uuid"
)

type ShortenURL struct {
	ID            string
	UserID        uuid.UUID
	OriginalValue string
	CreatedAt     time.Time
	Deleted       bool
}

func NewShortenURL(userID uuid.UUID, id, value string) *ShortenURL {
	return &ShortenURL{
		ID:            id,
		UserID:        userID,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
