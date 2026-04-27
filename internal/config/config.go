package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   []byte
	DefaultPort string
	JWTIssuer   string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "authentication-system"
	}

	return &Config{
		DatabaseURL: os.Getenv("DatabaseURL"),
		JWTSecret:   jwtSecret,
		DefaultPort: os.Getenv("DefaultPort"),
		JWTIssuer:   jwtIssuer,
	}
}
