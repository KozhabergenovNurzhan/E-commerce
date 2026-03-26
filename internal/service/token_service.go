package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, role domain.Role) (*domain.AuthTokens, error)
	ValidateAccessToken(token string) (*Claims, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type tokenService struct {
	repo       repository.TokenRepository
	userRepo   repository.UserRepository
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenService(
	repo repository.TokenRepository,
	userRepo repository.UserRepository,
	secret string,
	accessTTL, refreshTTL time.Duration,
) TokenService {
	return &tokenService{
		repo:       repo,
		userRepo:   userRepo,
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, role domain.Role) (*domain.AuthTokens, error) {
	accessToken, err := s.generateAccessToken(userID, role)
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
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *tokenService) generateAccessToken(userID uuid.UUID, role domain.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *tokenService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperrors.ErrInternal
	}
	raw := hex.EncodeToString(b)

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
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

func (s *tokenService) ValidateAccessToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrUnauthorized
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperrors.ErrUnauthorized
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, apperrors.ErrUnauthorized
	}
	return claims, nil
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
