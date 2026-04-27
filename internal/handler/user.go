package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dcaiovinicius/authentication-system/internal/middleware"
	"github.com/dcaiovinicius/authentication-system/internal/repository"
	"github.com/google/uuid"
)

type userResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from context (set by JWT middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.userRepo.GetUserByID(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(userResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
	})
}