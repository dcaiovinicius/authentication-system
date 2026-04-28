package setup

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcaiovinicius/authentication-system/infra/database"
	"github.com/dcaiovinicius/authentication-system/infra/support"
	"github.com/dcaiovinicius/authentication-system/internal/config"
	"github.com/dcaiovinicius/authentication-system/internal/server"
)

func SetupIntegrationTest(t *testing.T) *http.Server {
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

func DoRequest(
	t *testing.T,
	server *http.Server,
	method string,
	path string,
	body []byte,
) (*httptest.ResponseRecorder, *http.Request) {

	t.Helper()

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req := httptest.NewRequest(method, path, reqBody)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()

	server.Handler.ServeHTTP(rec, req)

	return rec, req
}
