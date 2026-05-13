package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/newmersedez/urlshort/internal/model"
)

// HTTPAuditObserver реализует AuditObserver, отправляя события на удалённый HTTP-сервер.
// Запросы выполняются с таймаутом 5 секунд.
type HTTPAuditObserver struct {
	url    string
	client *http.Client
}

// NewHTTPAuditObserver создаёт HTTPAuditObserver, отправляющий события POST-запросом на url.
func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// OnAuditEvent сериализует событие в JSON и отправляет его POST-запросом.
// Возвращает ошибку при сетевых проблемах или если сервер вернул статус не из диапазона 2xx.
func (hao *HTTPAuditObserver) OnAuditEvent(event *model.AuditEvent) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	resp, err := hao.client.Post(hao.url, "application/json", bytes.NewReader(eventBytes))
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit server returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
