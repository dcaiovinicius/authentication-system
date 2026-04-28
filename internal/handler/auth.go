package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcaiovinicius/authentication-system/internal/auth"
	"github.com/dcaiovinicius/authentication-system/internal/middleware"
	"github.com/dcaiovinicius/authentication-system/internal/repository"
	"github.com/google/uuid"
)

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type AuthHandler struct {
	authService        *auth.AuthService
	refreshTokenRepo   repository.RefreshTokenRepository
	refreshTokenExpiry time.Duration
}

func NewAuthHandler(authService *auth.AuthService, refreshTokenRepo repository.RefreshTokenRepository) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		refreshTokenRepo:   refreshTokenRepo,
		refreshTokenExpiry: time.Hour * 24 * 7, // 7 days
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, refreshToken, err := h.authService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, refreshToken, err := h.authService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, refreshToken, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDStr := r.Context().Value(middleware.UserIDContextKey)
	if userIDStr == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	err = h.refreshTokenRepo.RevokeByUserID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to logout"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
