package tests

import (
	"strings"
	"testing"

	"github.com/dcaiovinicius/authentication-system/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "1234"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	if !auth.CheckPasswordHash(password, hash) {
		t.Fatal("expected password to match hash")
	}
}

func TestInvalidPassword(t *testing.T) {
	hash, err := auth.HashPassword("1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auth.CheckPasswordHash("wrong", hash) {
		t.Fatal("expected password mismatch")
	}
}

func TestPasswordLengthLimit(t *testing.T) {
	invalid := strings.Repeat("a", 73)

	_, err := auth.HashPassword(invalid)

	if err != auth.ErrPasswordLengthLimit {
		t.Fatalf("expected %v, got %v", auth.ErrPasswordLengthLimit, err)
	}
}
