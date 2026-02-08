package model

type ShortenURL struct {
	ID            string
	OriginalValue string
}

func NewShortenURL(key, value string) *ShortenURL {
	return &ShortenURL{
		ID:            key,
		OriginalValue: value,
	}
}
