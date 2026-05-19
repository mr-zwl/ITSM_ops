package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"sub"`
	IssuedAt int64  `json:"iat"`
	ExpireAt int64  `json:"exp"`
}

func GenerateToken(secret string, userID uint64, username string, expire time.Duration) (string, error) {
	header := b64Encode([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := Claims{
		UserID:   userID,
		Username: username,
		IssuedAt: time.Now().Unix(),
		ExpireAt: time.Now().Add(expire).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := b64Encode(payload)
	sigInput := header + "." + encoded
	return sigInput + "." + hmacSign(sigInput, secret), nil
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	sigInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(hmacSign(sigInput, secret))) {
		return nil, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() > c.ExpireAt {
		return nil, ErrExpiredToken
	}
	return &c, nil
}

func hmacSign(input, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(input))
	return b64Encode(h.Sum(nil))
}

func b64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
