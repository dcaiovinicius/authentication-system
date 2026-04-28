package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   []byte
	DefaultPort string
	JWTIssuer   string
	RootPath    string
	Environment string
}

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func LoadConfig() *Config {
	root := findProjectRoot()
	envPath := filepath.Join(root, ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("No .env file found, relying on environment variables: %v", err)
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	jwtIssuer := os.Getenv("JWT_ISSUER")

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   jwtSecret,
		DefaultPort: os.Getenv("DefaultPort"),
		JWTIssuer:   jwtIssuer,
		RootPath:    root,
		Environment: os.Getenv("ENV"),
	}
}
