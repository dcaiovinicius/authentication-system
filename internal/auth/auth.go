package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/dcaiovinicius/authentication-system/internal/model"
	"github.com/dcaiovinicius/authentication-system/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidInput      = errors.New("invalid input")
)

type AuthService struct {
	userRepository    repository.UserRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	jwtSecret         []byte
	accessTokenExpiry time.Duration
	issuer            string
}

func NewAuthService(userRepository repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, jwtSecret []byte, issuer string) *AuthService {
	return &AuthService{
		userRepository:    userRepository,
		refreshTokenRepo:  refreshTokenRepo,
		jwtSecret:         jwtSecret,
		accessTokenExpiry: time.Hour * 24, // 24 hours
		issuer:            issuer,
	}
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) createAccessToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": now.Add(s.accessTokenExpiry).Unix(),
		"iat": now.Unix(),
		"iss": s.issuer,
		"aud": "authentication-system",
		"nbf": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) createRefreshToken(ctx context.Context, user *model.User) (string, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Hour * 24 * 7) // 7 days
	token := model.NewRefreshToken(user.ID, refreshToken, expiresAt)

	err = s.refreshTokenRepo.CreateRefreshToken(ctx, token)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, string, error) {
	// Validate input
	if email == "" || password == "" {
		return "", "", ErrInvalidInput
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return "", "", ErrInvalidToken
	}

	accessToken, err := s.createAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.createRefreshToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Register(ctx context.Context, email, username, password string) (string, string, error) {
	// Validate input
	if email == "" || username == "" || password == "" {
		return "", "", ErrInvalidInput
	}

	// Basic email validation
	if !strings.Contains(email, "@") {
		return "", "", ErrInvalidInput
	}

	// Password strength validation (minimum 8 characters)
	if len(password) < 8 {
		return "", "", ErrInvalidInput
	}

	existing, err := s.userRepository.GetUserByEmail(ctx, email)
	if err == nil && existing != nil {
		return "", "", ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Higher cost factor
	if err != nil {
		return "", "", err
	}

	user, err := model.NewUser(email, username, string(hash))
	if err != nil {
		return "", "", err
	}

	err = s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := s.Authenticate(ctx, email, password)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	token, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	if token.Revoked {
		return "", "", ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return "", "", ErrTokenExpired
	}

	user, err := s.userRepository.GetUserByID(ctx, token.UserID)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Revoke old refresh token
	err = s.refreshTokenRepo.RevokeByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	// Create new tokens
	accessToken, err := s.createAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.createRefreshToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
