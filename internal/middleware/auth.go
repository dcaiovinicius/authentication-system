package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dcaiovinicius/authentication-system/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

var ErrUnauthorized = errors.New("unauthorized")

type AuthMiddleware struct {
	jwtSecret []byte
	issuer    string
}

func NewAuthMiddleware(jwtSecret []byte, issuer string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret, issuer: issuer}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, auth.ErrInvalidToken
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		// Validate issuer
		if iss, ok := claims["iss"].(string); ok && iss != m.issuer {
			http.Error(w, "invalid token issuer", http.StatusUnauthorized)
			return
		}

		// Extract user ID from claims
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			http.Error(w, "invalid token subject", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, sub)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}
