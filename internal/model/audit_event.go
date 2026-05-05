package model

import (
	"encoding/json"
	"time"
)

// AuditEvent описывает факт выполнения операции над URL.
// Используется для аудит-лога: сохранения в файл или отправки на внешний сервер.
type AuditEvent struct {
	// Timestamp — Unix-время возникновения события.
	Timestamp int64 `json:"ts"`
	// Action — тип операции, например "shorten" или "follow".
	Action string `json:"action"`
	// UserID — строковое представление UUID пользователя, выполнившего действие.
	UserID string `json:"user_id"`
	// URL — исходный полный URL, к которому относится событие.
	URL string `json:"url"`
}

// NewAuditEvent создаёт AuditEvent с текущим Unix-временем.
func NewAuditEvent(action string, userID string, url string) *AuditEvent {
	return &AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}
}

// MarshalJSON сериализует AuditEvent в JSON.
func (ae *AuditEvent) MarshalJSON() ([]byte, error) {
	type Alias AuditEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(ae),
	})
}
