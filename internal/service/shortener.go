package service

import (
	"crypto/md5"
	"encoding/hex"
	urltools "net/url"
)

type Shortener struct{}

func NewURLShortenerService() *Shortener {
	return &Shortener{}
}

func (s *Shortener) Shorten(url string) (string, error) {
	if _, err := urltools.ParseRequestURI(url); err != nil {
		return "", err
	}

	hash := md5.Sum([]byte(url))
	encoded := hex.EncodeToString(hash[:])[:8]
	return encoded, nil
}
