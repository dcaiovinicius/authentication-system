package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    *time.Time
}

var (
	ErrEmptyEmail        = errors.New("email cannot be empty")
	ErrEmptyUsername     = errors.New("username cannot be empty")
	ErrEmptyPasswordHash = errors.New("password hash cannot be empty")
)

func NewUser(email, username, passwordHash string) (*User, error) {
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if username == "" {
		return nil, ErrEmptyUsername
	}
	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	return &User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	}, nil
}
