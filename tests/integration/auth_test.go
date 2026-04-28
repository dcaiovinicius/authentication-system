package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcaiovinicius/authentication-system/infra/database"
	"github.com/dcaiovinicius/authentication-system/infra/support"
	"github.com/dcaiovinicius/authentication-system/internal/config"
	"github.com/dcaiovinicius/authentication-system/internal/server"
)

type RegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func setupIntegrationTest(t *testing.T) *http.Server {
	t.Helper()

	support.RunUpMigrations()

	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		support.RunDownMigrations()
	})

	return server.NewServer(cfg, db)
}

func TestRegisterNewUser(t *testing.T) {
	server := setupIntegrationTest(t)

	t.Run("should register a new user", func(t *testing.T) {

		payload := map[string]string{
			"username": "testing",
			"email":    "testing@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		server.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp RegisterResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid response format: %v", err)
		}

		if resp.AccessToken == "" {
			t.Fatal("access token is empty")
		}

		if resp.RefreshToken == "" {
			t.Fatal("refresh token is empty")
		}
	})

	t.Run("should not register with existing email", func(t *testing.T) {

		payload := map[string]string{
			"username": "testing",
			"email":    "testing@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		server.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid response format: %v", err)
		}

		if resp["error"] == "" {
			t.Fatal("error message is empty")
		}
	})

	t.Run("should be able to login with registered user", func(t *testing.T) {
		payload := map[string]string{
			"username": "testing",
			"email":    "testing@example.com",
			"password": "password123",
		}

		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		server.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp LoginResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid response format: %v", err)
		}

		if resp.AccessToken == "" {
			t.Fatal("access token is empty")
		}

		if resp.RefreshToken == "" {
			t.Fatal("refresh token is empty")
		}
	})
}
