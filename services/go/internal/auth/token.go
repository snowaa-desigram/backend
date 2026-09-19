package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	JWTIssuer = "desigram-auth"
	// ClaimUserID — кастомный claim с id пользователя: стандартный `sub` JWT-middleware go-zero в контекст не кладёт.
	ClaimUserID = "uid"
	ClaimEmail  = "email"
)

// TokenIssuer выпускает access-JWT (HS256) — тот же секрет проверяют middleware go-zero и core.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenIssuer(secret string, ttl time.Duration, now func() time.Time) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl, now: now}
}

func (t *TokenIssuer) TTL() time.Duration { return t.ttl }

func (t *TokenIssuer) IssueAccess(u *User) (string, error) {
	now := t.now()
	claims := jwt.MapClaims{
		"iss":       JWTIssuer,
		"sub":       u.ID,
		"iat":       now.Unix(),
		"exp":       now.Add(t.ttl).Unix(),
		ClaimUserID: u.ID,
		ClaimEmail:  u.Email,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
}

// NewRefreshToken — opaque-токен: клиенту уходит raw, в БД хранится sha256.
func NewRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashToken(raw), nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// NewCode — 6-значный код подтверждения.
func NewCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

func HashCode(code string) string { return HashToken(code) }

func hashEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
