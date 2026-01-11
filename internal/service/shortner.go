package service

import (
	"crypto/md5"
	"encoding/hex"
	urltools "net/url"
)

type URLShortenerService struct{}

func NewURLShortenerService() *URLShortenerService {
	return new(URLShortenerService)
}

func (s *URLShortenerService) Shorten(url string) (*string, error) {
	_, err := urltools.ParseRequestURI(url)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum([]byte(url))
	encoded := hex.EncodeToString(hash[:])[:8]
	return &encoded, nil
}
