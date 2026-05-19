package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	hash := sha256.Sum256([]byte(saltHex + password))
	return fmt.Sprintf("%s:%s", saltHex, hex.EncodeToString(hash[:])), nil
}

func CheckPassword(hashed, password string) bool {
	parts := strings.SplitN(hashed, ":", 2)
	if len(parts) != 2 {
		return false
	}
	hash := sha256.Sum256([]byte(parts[0] + password))
	return hex.EncodeToString(hash[:]) == parts[1]
}
