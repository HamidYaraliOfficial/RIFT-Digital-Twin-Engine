package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// Claims are the payload of a RIFT session token.
type Claims struct {
	UserID    string    `json:"userId"`
	OrgID     string    `json:"orgId"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// TokenIssuer signs and verifies compact "<payload>.<signature>" tokens with
// HMAC-SHA256 — functionally equivalent to a JWT's HS256 mode without
// pulling in an external JWT library.
type TokenIssuer struct {
	secret []byte
}

func NewTokenIssuer(secret string) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret)}
}

func (t *TokenIssuer) Issue(c Claims, ttl time.Duration) (string, error) {
	c.IssuedAt = time.Now().UTC()
	c.ExpiresAt = c.IssuedAt.Add(ttl)
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	sig := t.sign(encoded)
	return encoded + "." + sig, nil
}

func (t *TokenIssuer) sign(encodedPayload string) string {
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(encodedPayload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (t *TokenIssuer) Verify(token string) (Claims, error) {
	var claims Claims
	sepIdx := -1
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			sepIdx = i
			break
		}
	}
	if sepIdx < 0 {
		return claims, errors.New("malformed token")
	}
	encoded, sig := token[:sepIdx], token[sepIdx+1:]
	if !hmac.Equal([]byte(t.sign(encoded)), []byte(sig)) {
		return claims, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return claims, err
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, err
	}
	if time.Now().UTC().After(claims.ExpiresAt) {
		return claims, errors.New("token expired")
	}
	return claims, nil
}
