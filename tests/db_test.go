package tests

import (
	"os"
	"testing"

	"github.com/dcaiovinicius/authentication-system/infra/database"
)

func TestDatabase_NoEnv(t *testing.T) {
	os.Unsetenv("DATABASE_URL")

	_, err := database.Connect()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "DATABASE_URL not set" {
		t.Fatalf("expected %q, got %q", "DATABASE_URL not set", err.Error())
	}
}
