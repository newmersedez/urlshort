package service

import (
	"log/slog"

	"github.com/newmersedez/urlshort/internal/model"
)

// AuditObserver описывает подписчика на аудит-события.
// Реализации могут записывать события в файл, отправлять на удалённый сервер и т. д.
type AuditObserver interface {
	// OnAuditEvent вызывается при возникновении аудит-события.
	// Возвращает ошибку, если доставка события не удалась.
	OnAuditEvent(event *model.AuditEvent) error
}

// AuditService уведомляет всех зарегистрированных наблюдателей об аудит-событиях.
// Каждый наблюдатель вызывается в отдельной горутине - доставка асинхронная.
type AuditService struct {
	logger    *slog.Logger
	observers []AuditObserver
}

// NewAuditService создаёт новый AuditService с переданным логгером.
func NewAuditService(logger *slog.Logger) *AuditService {
	return &AuditService{
		logger:    logger,
		observers: make([]AuditObserver, 0),
	}
}

// Subscribe добавляет нового наблюдателя. Nil-наблюдатели игнорируются.
func (as *AuditService) Subscribe(observer AuditObserver) {
	if observer != nil {
		as.observers = append(as.observers, observer)
	}
}

// Notify асинхронно отправляет событие всем зарегистрированным наблюдателям.
// Ошибки доставки логируются на уровне Warn и не прерывают обработку остальных подписчиков.
func (as *AuditService) Notify(event *model.AuditEvent) {
	for _, observer := range as.observers {
		go func(obs AuditObserver, evt *model.AuditEvent) {
			if err := obs.OnAuditEvent(evt); err != nil {
				as.logger.Warn("failed to notify audit observer about avent", "error", err)
			}
		}(observer, event)
	}
}
