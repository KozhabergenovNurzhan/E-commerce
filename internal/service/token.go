package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type TokenService struct {
	db         *sqlx.DB
	repo       repository.TokenRepository
	userRepo   repository.UserRepository
	authMgr    auth.Manager
	refreshTTL time.Duration
}

func NewTokenService(
	db *sqlx.DB,
	repo repository.TokenRepository,
	userRepo repository.UserRepository,
	authMgr auth.Manager,
	refreshTTL time.Duration,
) *TokenService {
	return &TokenService{
		db:         db,
		repo:       repo,
		userRepo:   userRepo,
		authMgr:    authMgr,
		refreshTTL: refreshTTL,
	}
}

func (s *TokenService) GenerateTokenPair(ctx context.Context, userID int64, role models.Role) (*models.AuthTokens, error) {
	accessToken, err := s.authMgr.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, apperrors.Internal("token generation failed", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.authMgr.AccessTTL().Seconds()),
	}, nil
}

func (s *TokenService) generateRefreshToken(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperrors.Internal("token generation failed", err)
	}
	raw := hex.EncodeToString(b)

	rt := &models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: utils.Now().Add(s.refreshTTL),
		Revoked:   false,
		CreatedAt: utils.Now(),
	}

	if err := s.repo.Save(ctx, rt); err != nil {
		return "", err
	}

	return raw, nil
}

// Refresh atomically revokes the old token and issues a new pair.
func (s *TokenService) Refresh(ctx context.Context, refreshToken string) (*models.AuthTokens, error) {
	hash := hashToken(refreshToken)

	rt, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid or expired token", nil)
	}
	if rt.Revoked || utils.Now().After(rt.ExpiresAt) {
		return nil, apperrors.Unauthorized("invalid or expired token", nil)
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	if s.db == nil {
		// test path: no real DB, use repo methods directly
		if err := s.repo.Revoke(ctx, hash); err != nil {
			return nil, err
		}
		return s.GenerateTokenPair(ctx, user.ID, user.Role)
	}

	// Generate the new raw token before starting the transaction.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, apperrors.Internal("token generation failed", err)
	}
	raw := hex.EncodeToString(b)
	newRT := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		ExpiresAt: utils.Now().Add(s.refreshTTL),
		Revoked:   false,
		CreatedAt: utils.Now(),
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`, hash); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	rows, err := sqlx.NamedQueryContext(ctx, tx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at)
		 VALUES (:user_id, :token_hash, :expires_at, :revoked, :created_at)
		 RETURNING id`, newRT)
	if err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&newRT.ID)
	}
	rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	accessToken, err := s.authMgr.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, apperrors.Internal("token generation failed", err)
	}

	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: raw,
		ExpiresIn:    int64(s.authMgr.AccessTTL().Seconds()),
	}, nil
}

func (s *TokenService) Revoke(ctx context.Context, refreshToken string) error {
	return s.repo.Revoke(ctx, hashToken(refreshToken))
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
