package service

import (
	"crypto/md5"
	"encoding/hex"
	urltools "net/url"
)

type UrlShortenerService struct {}

func NewUrlShortenerService() *UrlShortenerService {
	return new(UrlShortenerService);
}

func (s *UrlShortenerService) Shorten(url string) (*string, error) {
	_, err := urltools.ParseRequestURI(url)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum([]byte(url))
    encoded := hex.EncodeToString(hash[:])[:8]
    return &encoded, nil
}

