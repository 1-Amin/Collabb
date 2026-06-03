package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignAndParseToken(t *testing.T) {
	id := "user-123"
	token, err := SignToken(id)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != id {
		t.Errorf("want %s, got %s", id, claims.UserID)
	}
}

func TestParseExpiredToken(t *testing.T) {
	claims := Claims{
		UserID: "user-456",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
	if _, err := ParseToken(token); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestParseInvalidSignature(t *testing.T) {
	claims := Claims{UserID: "x"}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	if _, err := ParseToken(token); err == nil {
		t.Error("expected error for wrong secret")
	}
}
