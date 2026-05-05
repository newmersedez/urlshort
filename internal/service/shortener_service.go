// Package service содержит бизнес-логику сервиса сокращения URL:
// генерацию коротких идентификаторов, управление токенами аутентификации,
// запись аудит-событий и фоновое удаление помеченных ссылок.
package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// ShortenerService генерирует короткие идентификаторы для URL на основе MD5-хэша.
type ShortenerService struct{}

// NewURLShortenerService создаёт новый экземпляр ShortenerService.
func NewURLShortenerService() *ShortenerService {
	return &ShortenerService{}
}

// Shorten возвращает 8-символьный идентификатор для переданного URL.
// Ошибка возвращается, если URL не начинается с http:// или https://.
func (s *ShortenerService) Shorten(url string) (string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("failed to shorten URL %s: invalid scheme", url)
	}

	hash := md5.Sum([]byte(url))
	// encode only first 4 bytes → 8-char hex string, avoids allocating a full 32-char string
	encoded := hex.EncodeToString(hash[:4])
	return encoded, nil
}
