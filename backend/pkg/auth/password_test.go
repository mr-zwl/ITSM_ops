package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hashed, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPassword(hashed, "mypassword") {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword(hashed, "wrongpassword") {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPassword_Unique(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Error("two hashes of same password should differ (random salt)")
	}
}

func TestCheckPassword_BadFormat(t *testing.T) {
	if CheckPassword("nocolon", "test") {
		t.Error("should return false for bad hash format")
	}
}
