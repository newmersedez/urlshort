package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/newmersedez/urlshort/internal/model"
)

type FileAuditObserver struct {
	filePath string
	mu       sync.Mutex
}

func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		filePath: filePath,
	}
}

func (fao *FileAuditObserver) OnAuditEvent(event *model.AuditEvent) error {
	fao.mu.Lock()
	defer fao.mu.Unlock()

	file, err := os.OpenFile(fao.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	_, err = file.WriteString(string(eventBytes) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}
