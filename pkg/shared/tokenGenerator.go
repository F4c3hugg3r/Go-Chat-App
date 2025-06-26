package shared

import (
	"crypto/rand"
	"encoding/base64"
)

// generateSecureToken generates a token containing random chars
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
