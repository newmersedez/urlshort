// Package repository предоставляет фабричные функции для создания хранилищ.
// Выбор реализации (in-memory, файл, PostgreSQL) производится в main на основе конфигурации.
package repository

import (
	"context"
	"log/slog"

	"github.com/newmersedez/urlshort/internal/repository/db"
	"github.com/newmersedez/urlshort/internal/repository/file"
	"github.com/newmersedez/urlshort/internal/repository/memory"
)

// NewMemoryRepository создаёт in-memory хранилище.
// Данные хранятся только в оперативной памяти и не сохраняются при перезапуске.
func NewMemoryRepository() (*memory.MemoryRepository, error) {
	return memory.NewMemoryRepository()
}

// NewFileRepository создаёт файловое хранилище по пути filepath.
// При запуске восстанавливает ранее сохранённые данные из JSONL-файла.
func NewFileRepository(filepath string) (*file.FileRepository, error) {
	return file.NewFileRepository(filepath)
}

// NewDBRepository создаёт PostgreSQL-хранилище по строке подключения databaseDSN.
// Перед созданием пула соединений автоматически выполняются миграции схемы.
func NewDBRepository(ctx context.Context, databaseDSN string, logger *slog.Logger) (*db.DBRepository, error) {
	return db.NewDBRepository(ctx, databaseDSN, logger)
}
