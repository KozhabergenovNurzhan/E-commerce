package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID int64, role domain.Role) (*domain.AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type tokenService struct {
	repo       repository.TokenRepository
	userRepo   repository.UserRepository
	authMgr    auth.Manager
	refreshTTL time.Duration
}

func NewTokenService(
	repo repository.TokenRepository,
	userRepo repository.UserRepository,
	authMgr auth.Manager,
	refreshTTL time.Duration,
) TokenService {
	return &tokenService{
		repo:       repo,
		userRepo:   userRepo,
		authMgr:    authMgr,
		refreshTTL: refreshTTL,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, userID int64, role domain.Role) (*domain.AuthTokens, error) {
	accessToken, err := s.authMgr.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	refreshToken, err := s.generateRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.authMgr.AccessTTL().Seconds()),
	}, nil
}

func (s *tokenService) generateRefreshToken(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperrors.ErrInternal
	}
	raw := hex.EncodeToString(b)

	rt := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().Add(s.refreshTTL),
		Revoked:   false,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, rt); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *tokenService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	hash := hashToken(refreshToken)

	rt, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, apperrors.ErrUnauthorized
	}

	// Rotate: revoke old token before issuing new pair.
	if err := s.repo.Revoke(ctx, hash); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(ctx, user.ID, user.Role)
}

func (s *tokenService) Revoke(ctx context.Context, refreshToken string) error {
	return s.repo.Revoke(ctx, hashToken(refreshToken))
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
