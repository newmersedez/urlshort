package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewShortenURL(t *testing.T) {
	// Arrange
	key, value := "key", "value"
	userID := uuid.New()

	// Act
	shortenURL := NewShortenURL(userID, key, value)

	// Assert
	require.Equal(t, key, shortenURL.ID)
	require.Equal(t, userID, shortenURL.UserID)
	require.Equal(t, value, shortenURL.OriginalValue)
	require.False(t, shortenURL.Deleted)
}
