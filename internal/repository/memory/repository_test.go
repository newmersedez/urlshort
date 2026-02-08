package memory

import (
	"testing"

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
	shortenURL := model.NewShortenURL(key, value)

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
	shortenURL := model.NewShortenURL(key, value)

	repository, _ := NewMemoryRepository()
	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL, val)
}
