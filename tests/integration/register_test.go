package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/dcaiovinicius/authentication-system/internal/handler"
	"github.com/dcaiovinicius/authentication-system/tests/setup"
)

func TestRegisterNewUser(t *testing.T) {
	server := setup.SetupIntegrationTest(t)

	t.Run("should register a new user", func(t *testing.T) {
		rec, _ := setup.DoRequest(t, server, http.MethodPost, "/api/v1/register", NewRegisterRequestPayload(t))

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp handler.AuthResponse
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

		rec, _ := setup.DoRequest(t, server, http.MethodPost, "/api/v1/register", NewRegisterRequestPayload(t))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var resp handler.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid response format: %v", err)
		}

		if resp.Error == "" {
			t.Fatal("error message is empty")
		}
	})
}

func NewRegisterRequestPayload(t *testing.T) []byte {
	t.Helper()

	payload := handler.RegisterRequest{
		Username: "testing",
		Email:    "testing@example.com",
		Password: "password123",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("invalid payload: %v", err)
	}

	return body
}
