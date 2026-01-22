package repository

import (
	"testing"

	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	// Arrange

	// Act
	repository := NewRepository()

	// Assert
	require.NotNil(t, repository.data)
}

func TestAdd(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(key, value)

	repository := NewRepository()

	// Act
	repository.Add(t.Context(), shortenURL)

	// Assert
	val, ok := repository.data[key]
	require.True(t, ok)
	require.Equal(t, value, val.Value)
}


func TestGet(t *testing.T) {
	// Arrange
	key := "key"
	value := "value"
	shortenURL := model.NewShortenURL(key, value)

	repository := NewRepository()
	repository.Add(t.Context(), shortenURL)

	// Act
	val, err := repository.Get(t.Context(), key)

	// Assert
	require.NoError(t, err)
	require.Equal(t, shortenURL.Value, val.Value)
}