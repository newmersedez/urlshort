package service

import (
	"github.com/newmersedez/urlshort/internal/model"
	"log/slog"
)

type AuditObserver interface {
	OnAuditEvent(event *model.AuditEvent) error
}

type AuditService struct {
	logger	  *slog.Logger
	observers []AuditObserver
}

func NewAuditService(logger *slog.Logger) *AuditService {
	return &AuditService{
		logger: logger,
		observers: make([]AuditObserver, 0),
	}
}

func (as *AuditService) Subscribe(observer AuditObserver) {
	if observer != nil {
		as.observers = append(as.observers, observer)
	}
}

func (as *AuditService) Notify(event *model.AuditEvent) {
	for _, observer := range as.observers {
		go func(obs AuditObserver, evt *model.AuditEvent) {
			if err := obs.OnAuditEvent(evt); err != nil {
				as.logger.Warn("failed to notify audit observer about avent", "error", err)
			}
		}(observer, event)
	}
}