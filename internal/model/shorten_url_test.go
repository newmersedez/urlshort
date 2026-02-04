package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewShortenURL(t *testing.T) {
	// Arrange
	key, value := "key", "value"

	// Act
	shortenURL := NewShortenURL(key, value)

	// Assert
	require.Equal(t, key, shortenURL.Key)
	require.Equal(t, value, shortenURL.Value)
}
