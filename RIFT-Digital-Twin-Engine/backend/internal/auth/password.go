// Package auth implements user management, credential storage and signed
// session tokens (RBAC/ABAC + the Twin Permission Model). It has no external
// dependency: passwords are salted and stretched with a manual PBKDF2-style
// loop over crypto/sha256, and sessions are HMAC-SHA256 signed tokens. This
// keeps the whole platform buildable offline; production deployments are
// expected to front this with real OAuth2/OIDC (see README) while keeping
// the same RBAC/ABAC model underneath.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const iterations = 100_000

// HashPassword returns (hash, salt) both hex-encoded.
func HashPassword(password string) (hash string, salt string) {
	saltBytes := make([]byte, 16)
	_, _ = rand.Read(saltBytes)
	salt = hex.EncodeToString(saltBytes)
	return derive(password, salt), salt
}

func VerifyPassword(password, hash, salt string) bool {
	return hmac.Equal([]byte(derive(password, salt)), []byte(hash))
}

func derive(password, salt string) string {
	data := []byte(salt + ":" + password)
	sum := sha256.Sum256(data)
	for i := 0; i < iterations; i++ {
		sum = sha256.Sum256(sum[:])
	}
	return hex.EncodeToString(sum[:])
}
