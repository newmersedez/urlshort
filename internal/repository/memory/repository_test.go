package memory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	// Arrange

	// Act
	repository, err := NewMemoryRepository()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, repository.items)

	require.NoError(t, err)
}

func TestAdd(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(uuid.New(), key, value)

	repository, _ := NewMemoryRepository()

	// Act
	repository.Add(t.Context(), shortenURL)

	// Assert
	val, ok := repository.items[key]
	require.True(t, ok)
	require.Equal(t, value, val.OriginalValue)
}

func TestGet(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(uuid.New(), key, value)

	repository, _ := NewMemoryRepository()
	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL, val)
}

func TestGetList(t *testing.T) {
	// Arrange
	userID := uuid.New()
	shortenURL := model.NewShortenURL(userID, "key1", "value1")
	anotherShortenURL := model.NewShortenURL(uuid.New(), "key2", "value2")

	repository, _ := NewMemoryRepository()
	repository.Add(t.Context(), shortenURL)
	repository.Add(t.Context(), anotherShortenURL)

	// Act
	val, err := repository.GetList(t.Context(), userID)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, len(val))
	require.Equal(t, shortenURL.ID, val[0].ID)
	require.Equal(t, shortenURL.UserID, val[0].UserID)
	require.Equal(t, shortenURL.OriginalValue, val[0].OriginalValue)
	require.Equal(t, shortenURL.CreatedAt, val[0].CreatedAt)
}
