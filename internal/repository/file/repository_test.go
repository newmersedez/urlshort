package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	// Arrange
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	// Act
	repository, err := NewFileRepository(fileStoragePath)

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
	userID := uuid.New()
	shortenURL := model.NewShortenURL(userID, key, value)
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	repository, _ := NewFileRepository(fileStoragePath)
	defer os.Remove(fileStoragePath)

	// Act
	repository.Add(t.Context(), shortenURL)

	// Assert
	val, ok := repository.items[key]
	require.True(t, ok)
	require.Equal(t, userID, val.UserID)
	require.Equal(t, value, val.OriginalValue)
}

func TestGet(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	userID := uuid.New()
	shortenURL := model.NewShortenURL(userID, key, value)
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	repository, _ := NewFileRepository(fileStoragePath)
	defer os.Remove(fileStoragePath)

	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL.UserID, val.UserID)
	require.Equal(t, shortenURL.OriginalValue, val.OriginalValue)
}

func TestGetList(t *testing.T) {
	// Arrange
	userID := uuid.New()
	shortenURL := model.NewShortenURL(userID, "key1", "value2")
	fileStoragePath := filepath.Join(os.TempDir(), "test*.json")

	repository, _ := NewFileRepository(fileStoragePath)
	defer os.Remove(fileStoragePath)

	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.GetList(t.Context(), userID)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, len(val))
	require.Equal(t, shortenURL.ID, val[0].ID)
	require.Equal(t, shortenURL.UserID, val[0].UserID)
	require.Equal(t, shortenURL.OriginalValue, val[0].OriginalValue)
}
