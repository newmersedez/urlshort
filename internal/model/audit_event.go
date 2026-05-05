package model

import (
	"encoding/json"
	"time"
)

type AuditEvent struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}

func NewAuditEvent(action string, userID string, url string) *AuditEvent {
	return &AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}
}

func (ae *AuditEvent) MarshalJSON() ([]byte, error) {
	type Alias AuditEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(ae),
	})
}
