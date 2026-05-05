package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

type ShortenerService struct{}

func NewURLShortenerService() *ShortenerService {
	return &ShortenerService{}
}

func (s *ShortenerService) Shorten(url string) (string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("failed to shorten URL %s: invalid scheme", url)
	}

	hash := md5.Sum([]byte(url))
	encoded := hex.EncodeToString(hash[:4])
	return encoded, nil
}
