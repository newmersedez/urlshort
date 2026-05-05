package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/newmersedez/urlshort/internal/model"
)

type HTTPAuditObserver struct {
	url    string
	client *http.Client
}

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

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
