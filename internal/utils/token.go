package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// GenerateResetToken creates a 256-bit secret independent of login JWTs.
func GenerateResetToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func HashResetToken(token string) (string, error) {
	if len(token) != 64 {
		return "", errors.New("invalid reset token")
	}
	raw, err := hex.DecodeString(token)
	if err != nil || len(raw) != 32 {
		return "", errors.New("invalid reset token")
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:]), nil
}
