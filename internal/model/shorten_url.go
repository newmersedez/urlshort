// Package model содержит доменные модели сервиса сокращения URL.
package model

import (
	"time"

	"github.com/google/uuid"
)

// ShortenURL представляет сокращённую ссылку, хранящуюся в системе.
type ShortenURL struct {
	// ID - короткий идентификатор (8 hex-символов).
	ID string
	// UserID - идентификатор пользователя-владельца ссылки.
	UserID uuid.UUID
	// OriginalValue - исходный полный URL.
	OriginalValue string
	// CreatedAt - время создания записи в UTC.
	CreatedAt time.Time
	// Deleted - признак мягкого удаления; true означает, что ссылка деактивирована.
	Deleted bool
}

// NewShortenURL создаёт новую запись ShortenURL с текущим временем создания.
func NewShortenURL(userID uuid.UUID, id, value string) *ShortenURL {
	return &ShortenURL{
		ID:            id,
		UserID:        userID,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
