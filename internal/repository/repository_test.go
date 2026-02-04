package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	// Arrange
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	// Act
	repository, err := NewRepository(fileStoragePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, repository.items)
	require.NotNil(t, repository.encoder)
	require.NotNil(t, repository.file)

	_, err = os.Stat(fileStoragePath)
	require.NoError(t, err)
}

func TestAdd(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(key, value)
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	repository, _ := NewRepository(fileStoragePath)
	defer os.Remove(fileStoragePath)

	// Act
	repository.Add(t.Context(), shortenURL)

	// Assert
	val, ok := repository.items[key]
	require.True(t, ok)
	require.Equal(t, value, val.Value)
}


func TestGet(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(key, value)
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	repository, _ := NewRepository(fileStoragePath)
	defer os.Remove(fileStoragePath)

	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL.Value, val.Value)
}