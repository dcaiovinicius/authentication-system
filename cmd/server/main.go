package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dcaiovinicius/authentication-system/infra/database"
	"github.com/dcaiovinicius/authentication-system/internal/auth"
	"github.com/dcaiovinicius/authentication-system/internal/config"
	"github.com/dcaiovinicius/authentication-system/internal/handler"
	"github.com/dcaiovinicius/authentication-system/internal/middleware"
	"github.com/dcaiovinicius/authentication-system/internal/repository"
)

func main() {
	cfg := config.NewConfig()

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()

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

	go func() {
		log.Println("server running on", cfg.DefaultPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	log.Println("server stopped cleanly")
}

// middleware simples (log + proteção básica)
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
