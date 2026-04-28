package database

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

	"github.com/dcaiovinicius/authentication-system/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.DatabaseURL

	if cfg.Environment == "test" {
		dsn = withTestDB(dsn)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func withTestDB(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("invalid dsn: %v", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	u.Path = "/" + dbName + "_test"

	return u.String()
}
