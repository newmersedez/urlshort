package model

type ShortenUrl struct {
	ShortenValue	string
	OriginalValue 	string
}

func NewShortenUrl(shortenValue string, originalValue string) *ShortenUrl{
	return &ShortenUrl{
		ShortenValue: shortenValue,
		OriginalValue: originalValue,
	}
}
