package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/dcaiovinicius/authentication-system/internal/auth"
	"github.com/dcaiovinicius/authentication-system/internal/config"
	"github.com/dcaiovinicius/authentication-system/internal/handler"
	"github.com/dcaiovinicius/authentication-system/internal/middleware"
	"github.com/dcaiovinicius/authentication-system/internal/repository"
)

func NewServer(cfg *config.Config, db *sql.DB) *http.Server {
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := auth.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret, cfg.JWTIssuer)
	authHandler := handler.NewAuthHandler(authService, refreshTokenRepo)
	userHandler := handler.NewUserHandler(userRepo)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, cfg.JWTIssuer)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", loggingMiddleware(middleware.CORS(authHandler.Register)))
	mux.HandleFunc("/api/v1/login", loggingMiddleware(middleware.CORS(authHandler.Login)))
	mux.HandleFunc("/api/v1/refresh", loggingMiddleware(middleware.CORS(authHandler.Refresh)))
	mux.HandleFunc("/api/v1/logout", loggingMiddleware(middleware.CORS(authMiddleware.Authenticate(authHandler.Logout))))
	mux.HandleFunc("/api/v1/user", loggingMiddleware(middleware.CORS(authMiddleware.Authenticate(userHandler.GetCurrentUser))))

	server := &http.Server{
		Addr:         cfg.DefaultPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				log.Printf("panic: %v", rec)
			}
		}()

		next(w, r)

		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}
