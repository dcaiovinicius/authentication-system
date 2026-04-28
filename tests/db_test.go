package tests

import (
	"testing"

	"github.com/dcaiovinicius/authentication-system/infra/database"
	"github.com/dcaiovinicius/authentication-system/internal/config"
)

func TestDatabaseConnection(t *testing.T) {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)

	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()
}
