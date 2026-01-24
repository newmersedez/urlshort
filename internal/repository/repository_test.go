package repository

import (
	"os"
	"testing"

	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	// Arrange
	fileStoragePath := "test.json"

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

	repository, _ := NewRepository("test.json")

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

	repository, _ := NewRepository("test.json")
	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL.Value, val.Value)
}