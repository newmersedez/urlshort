package model

type ShortenURL struct {
	Key  string
	Value string
}

func NewShortenURL(key, value string) *ShortenURL {
	return &ShortenURL{
		Key:  key,
		Value: value,
	}
}
