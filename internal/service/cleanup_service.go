package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

// DeleteRequest описывает задачу на физическое удаление списка URL конкретного пользователя.
type DeleteRequest struct {
	UserID        uuid.UUID
	ShortenURLIDs []string
}

// Repository описывает интерфейс хранилища, необходимый CleanupService.
type Repository interface {
	// Get возвращает сокращённый URL по идентификатору.
	Get(ctx context.Context, id string) (*model.ShortenURL, error)
	// GetList возвращает активные URL пользователя.
	GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	// GetDeletedList возвращает мягко удалённые URL для восстановления очереди при старте.
	GetDeletedList(ctx context.Context) ([]model.ShortenURL, error)
	// Add сохраняет новый URL.
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	// AddBatch сохраняет пакет URL.
	AddBatch(ctx context.Context, shortenURLs []*model.ShortenURL) error
	// SoftDeleteBatch помечает URL как удалённые.
	SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error
	// HardDeleteBatch физически удаляет URL.
	HardDeleteBatch(ctx context.Context, ids []string) error
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
	// Close освобождает ресурсы хранилища.
	Close()
}

// CleanupService собирает задачи на удаление URL в канал и периодически
// выполняет пакетное физическое удаление через Repository.HardDeleteBatch.
type CleanupService struct {
	queue  chan DeleteRequest
	store  Repository
	logger *slog.Logger
}

// NewCleanupService создаёт CleanupService и восстанавливает очередь удалений
// из мягко удалённых записей хранилища (на случай перезапуска).
func NewCleanupService(ctx context.Context, store Repository, logger *slog.Logger) *CleanupService {
	service := &CleanupService{
		queue:  make(chan DeleteRequest, 1024),
		store:  store,
		logger: logger,
	}

	service.loadPending(ctx)

	return service
}

// ScheduleDelete неблокирующим образом добавляет задачу удаления в очередь.
// Если контекст уже отменён до отправки, задача не добавляется и генерируется предупреждение.
func (s *CleanupService) ScheduleDelete(ctx context.Context, userID uuid.UUID, ids []string) {
	go func() {
		select {
		case <-ctx.Done():
			s.logger.Warn("context cancelled while scheduling delete", "ids_count", len(ids))
		case s.queue <- DeleteRequest{UserID: userID, ShortenURLIDs: ids}:
			s.logger.Debug("added hard deletion task to cleanup service", "ids_count", len(ids))
		}
	}()
}

// Start запускает фоновый цикл обработки очереди удалений.
// URL удаляются пакетами до 100 штук или по истечении 10-секундного интервала.
// Завершается при отмене ctx, выполняя финальный сброс буфера.
func (s *CleanupService) Start(ctx context.Context) {
	const batchSize = 100
	const flushInterval = 10 * time.Second

	buffer := make([]string, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) > 0 {
			if err := s.store.HardDeleteBatch(ctx, buffer); err != nil {
				s.logger.Error("failed to hard delete batch", "error", err)
			} else {
				s.logger.Debug("hard deleted URLs", "count", len(buffer))
			}
			buffer = buffer[:0]
		}
	}

	for {
		select {
		case req := <-s.queue:
			for _, id := range req.ShortenURLIDs {
				buffer = append(buffer, id)
				if len(buffer) >= batchSize {
					flush()
				}
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (s *CleanupService) loadPending(ctx context.Context) error {
	pendingUrls, err := s.store.GetDeletedList(ctx)
	if err != nil {
		return fmt.Errorf("failed to restore deletion tasks queue: %w", err)
	}

	userUrls := make(map[uuid.UUID][]string)

	for _, url := range pendingUrls {
		if _, exists := userUrls[url.UserID]; !exists {
			userUrls[url.UserID] = make([]string, 0)
		}
		userUrls[url.UserID] = append(userUrls[url.UserID], url.ID)
	}

	for userID, urlIDs := range userUrls {
		go func() {
			select {
			case <-ctx.Done():
				s.logger.Warn("context cancelled while scheduling delete", "ids_count", len(urlIDs))
			case s.queue <- DeleteRequest{UserID: userID, ShortenURLIDs: urlIDs}:
				s.logger.Debug("added hard deletion task to cleanup service", "ids_count", len(urlIDs))
			}
		}()
	}
	s.logger.Debug("restored hard deletion tasks queue", "urls_count", len(pendingUrls))

	return nil
}
