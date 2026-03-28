package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type TokenRepository interface {
	Save(ctx context.Context, t *models.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
}

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Save(ctx context.Context, t *models.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at)
		VALUES (:user_id, :token_hash, :expires_at, :revoked, :created_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, t)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&t.ID)
	}

	return nil
}

func (r *tokenRepository) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken

	const q = `SELECT * FROM refresh_tokens WHERE token_hash = $1`

	if err := r.db.GetContext(ctx, &t, q, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("token not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &t, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, hash string) error {
	const q = `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`

	if _, err := r.db.ExecContext(ctx, q, hash); err != nil {
		return apperrors.Internal("internal server error", err)
	}

	return nil
}
