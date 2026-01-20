package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	urltools "net/url"
)

type URLShortenerService struct{}

func NewURLShortenerService() *URLShortenerService {
	return &URLShortenerService{}
}

func (s *URLShortenerService) Shorten(url string) (string, error) {
	if _, err := urltools.ParseRequestURI(url); err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	hash := md5.Sum([]byte(url))
	encoded := hex.EncodeToString(hash[:])[:8]
	return encoded, nil
}
