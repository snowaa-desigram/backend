package auth_test

import (
	"testing"
	"time"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func init() { auth.SetBcryptCost(bcrypt.MinCost) } // тесты: bcrypt cost 12 слишком медленный под -race

// claimsOf проверяет подпись и возвращает claims; сроки не проверяются — часы в тестах фиктивные.
func claimsOf(t *testing.T, token string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	if _, err := parser.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) { return []byte(testSecret), nil }); err != nil {
		t.Fatalf("parse jwt: %v", err)
	}
	return claims
}

func TestIssueAccess(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	iss := auth.NewTokenIssuer(testSecret, 15*time.Minute, func() time.Time { return now })
	tok, err := iss.IssueAccess(&auth.User{ID: "u1", Email: testEmail})
	if err != nil {
		t.Fatal(err)
	}
	c := claimsOf(t, tok)
	if c["sub"] != "u1" || c[auth.ClaimUserID] != "u1" || c[auth.ClaimEmail] != testEmail || c["iss"] != auth.JWTIssuer {
		t.Fatalf("claims = %v", c)
	}
	if exp := int64(c["exp"].(float64)); exp != now.Add(15*time.Minute).Unix() {
		t.Fatalf("exp = %d", exp)
	}
	// чужой секрет — не проходит
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	if _, err := parser.Parse(tok, func(*jwt.Token) (any, error) { return []byte("other"), nil }); err == nil {
		t.Fatal("token verified with wrong secret")
	}
}

func TestRefreshAndCode(t *testing.T) {
	raw, hash, err := auth.NewRefreshToken()
	if err != nil || len(raw) < 40 || hash != auth.HashToken(raw) {
		t.Fatalf("raw=%q hash=%q err=%v", raw, hash, err)
	}
	raw2, _, _ := auth.NewRefreshToken()
	if raw == raw2 {
		t.Fatal("refresh tokens must be unique")
	}
	for i := 0; i < 100; i++ {
		code, err := auth.NewCode()
		if err != nil || !codeRegexp.MatchString(code) || len(code) != 6 {
			t.Fatalf("code=%q err=%v", code, err)
		}
	}
}

func TestPassword(t *testing.T) {
	h, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(h, testPassword) || auth.CheckPassword(h, "nope") {
		t.Fatal("bcrypt check failed")
	}
}
