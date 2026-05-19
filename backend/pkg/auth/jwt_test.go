package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key"
	token, err := GenerateToken(secret, 42, "admin", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("Username = %s, want admin", claims.Username)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken("secret-a", 1, "user", time.Hour)
	_, err := ParseToken("secret-b", token)
	if err != ErrInvalidToken {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_Expired(t *testing.T) {
	token, _ := GenerateToken("secret", 1, "user", -time.Hour)
	_, err := ParseToken("secret", token)
	if err != ErrExpiredToken {
		t.Errorf("err = %v, want ErrExpiredToken", err)
	}
}

func TestParseToken_Malformed(t *testing.T) {
	_, err := ParseToken("secret", "not.a.valid.token")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}
