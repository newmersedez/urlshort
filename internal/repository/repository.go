package repository

import (
	"database/sql"

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

func NewDBRepository(database *sql.DB) (*db.DBRepository, error) {
	return db.NewDBRepository(database)
}
