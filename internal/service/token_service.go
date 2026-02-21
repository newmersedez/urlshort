package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)
// 32 byte key
const secretKey = "27e655be04ff08f42dde5ecae167ac65d3c8e53e61afac06e607ad9fd4598bff"

type TokenService struct {
	aesgcm cipher.AEAD
}

func NewTokenService() (*TokenService, error) {
	keyBytes, err := hex.DecodeString(secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	aesblock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &TokenService{
		aesgcm: aesgcm,
	}, nil
}

func (a *TokenService) GetToken(userID uuid.UUID) (string, error) {
	nonce := make([]byte, a.aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	src := []byte(userID.String())
	encrypted := a.aesgcm.Seal(nil, nonce, src, nil)

	token := append(nonce, encrypted...)

	return hex.EncodeToString(token), nil
}

func (a *TokenService) GetUserId(token string) (uuid.UUID, error) {
	data, err := hex.DecodeString(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token format: %w", err)
	}

	nonceSize := a.aesgcm.NonceSize()
	if len(data) < nonceSize {
		return uuid.Nil, fmt.Errorf("token too short")
	}

	nonce := data[:nonceSize]
	encrypted := data[nonceSize:]

	decrypted, err := a.aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	userID, err := uuid.ParseBytes(decrypted)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid userID in token: %w", err)
	}

	return userID, nil
}

func (a *TokenService) IsValid(token string) bool {
	_, err := a.GetUserId(token)
	return err == nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
