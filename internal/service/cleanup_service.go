package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

type DeleteRequest struct {
	UserID uuid.UUID
	IDs    []string
}

type Repository interface {
	Get(ctx context.Context, id string) (*model.ShortenURL, error)
	GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	AddBatch(ctx context.Context, shortenURLs []*model.ShortenURL) error
	SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error
	HardDeleteBatch(ctx context.Context, ids []string) error
	Ping(ctx context.Context) error
	Close()
}

type CleanupService struct {
	queue  chan DeleteRequest
	store  Repository
	logger *slog.Logger
}

func NewCleanupService(store Repository, logger *slog.Logger) *CleanupService {
	return &CleanupService{
		queue:  make(chan DeleteRequest, 1024),
		store:  store,
		logger: logger,
	}
}

func (s *CleanupService) ScheduleDelete(ctx context.Context, userID uuid.UUID, ids []string) {
	go func() {
		select {
		case <-ctx.Done():
			s.logger.Warn("context cancelled while scheduling delete", "ids_count", len(ids))
		case s.queue <- DeleteRequest{UserID: userID, IDs: ids}:
			s.logger.Debug("added hard deletion task to cleanup service", "ids_count", len(ids))
		}
	}()
}

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
				s.logger.Info("hard deleted URLs", "count", len(buffer))
			}
			buffer = buffer[:0]
		}
	}

	for {
		select {
		case req := <-s.queue:
			for _, id := range req.IDs {
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