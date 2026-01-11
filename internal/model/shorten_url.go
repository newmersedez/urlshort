package model

type ShortenURL struct {
	ShortenValue  string
	OriginalValue string
}

func NewShortenURL(shortenValue string, originalValue string) *ShortenURL {
	return &ShortenURL{
		ShortenValue:  shortenValue,
		OriginalValue: originalValue,
	}
}
