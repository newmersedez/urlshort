package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	urltools "net/url"
)

type ShortenerService struct{}

func NewURLShortenerService() *ShortenerService {
	return &ShortenerService{}
}

func (s *ShortenerService) Shorten(url string) (string, error) {
	if _, err := urltools.ParseRequestURI(url); err != nil {
		return "", fmt.Errorf("failed to shorten URL %s: %w", url, err)
	}

	hash := md5.Sum([]byte(url))
	encoded := hex.EncodeToString(hash[:])[:8]
	return encoded, nil
}
