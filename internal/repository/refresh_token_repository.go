package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dcaiovinicius/authentication-system/internal/model"
	"github.com/google/uuid"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrCreateRefreshToken   = errors.New("create refresh token")
	ErrRevokeRefreshToken   = errors.New("revoke refresh token")
)

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, false)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return ErrCreateRefreshToken
	}

	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	row := r.db.QueryRowContext(ctx, query, token)
	return scanRefreshToken(row)
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return ErrRevokeRefreshToken
	}

	return nil
}

func scanRefreshToken(row *sql.Row) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}

	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.Token,
		&t.ExpiresAt,
		&t.CreatedAt,
		&t.Revoked,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return t, nil
}
