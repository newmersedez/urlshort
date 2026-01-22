package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewShortenUrl(t *testing.T) {
	// Arrange
	key, value := "key", "value"

	// Act
	shortenUrl := NewShortenURL(key, value)

	// Assert
	require.Equal(t, key, shortenUrl.Key)
	require.Equal(t, value, shortenUrl.Value)
}