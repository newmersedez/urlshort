package repository

import (
	"context"
	"log/slog"

	"github.com/newmersedez/urlshort/internal/repository/db"
	"github.com/newmersedez/urlshort/internal/repository/file"
	"github.com/newmersedez/urlshort/internal/repository/memory"
)

func NewMemoryRepository() (*memory.MemoryRepository, error) {
	return memory.NewMemoryRepository()
}

func NewFileRepository(filepath string) (*file.FileRepository, error) {
	return file.NewFileRepository(filepath)
}

func NewDBRepository(ctx context.Context, databaseDSN string, logger *slog.Logger) (*db.DBRepository, error) {
	return db.NewDBRepository(ctx, databaseDSN, logger)
}
