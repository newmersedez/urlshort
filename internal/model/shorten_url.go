package model

import "time"

type ShortenURL struct {
	ID            string
	UserID		  string
	OriginalValue string
	CreatedAt     time.Time
}

func NewShortenURL(userID, id, value string) *ShortenURL {
	return &ShortenURL{
		ID:            id,
		UserID: 	   userID,
		OriginalValue: value,
		CreatedAt:     time.Now().UTC(),
	}
}
