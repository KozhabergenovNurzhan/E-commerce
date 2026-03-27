package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

func TestGenerateTokenPair(t *testing.T) {
	tests := []struct {
		name          string
		generateFn    func(int64, domain.Role) (string, error)
		saveFn        func(context.Context, *domain.RefreshToken) error
		wantErr       error
		wantAccessTok string
	}{
		{
			name: "success",
			generateFn: func(_ int64, _ domain.Role) (string, error) {
				return "access-token", nil
			},
			saveFn: func(_ context.Context, rt *domain.RefreshToken) error {
				rt.ID = 1
				return nil
			},
			wantAccessTok: "access-token",
		},
		{
			name: "auth manager failure",
			generateFn: func(_ int64, _ domain.Role) (string, error) {
				return "", apperrors.ErrInternal
			},
			wantErr: apperrors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMgr := &testutil.MockAuthManager{
				GenerateAccessTokenFn: tt.generateFn,
				AccessTTLFn:           func() time.Duration { return 15 * time.Minute },
			}
			tokenRepo := &testutil.MockTokenRepo{SaveFn: tt.saveFn}

			svc := service.NewTokenService(tokenRepo, nil, authMgr, 7*24*time.Hour)
			tokens, err := svc.GenerateTokenPair(context.Background(), 1, domain.RoleCustomer)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAccessTok, tokens.AccessToken)
			assert.NotEmpty(t, tokens.RefreshToken)
			assert.Equal(t, int64(900), tokens.ExpiresIn)
		})
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name         string
		storedToken  *domain.RefreshToken
		findErr      error
		wantErr      error
		wantNewToken bool
	}{
		{
			name: "success",
			storedToken: &domain.RefreshToken{
				ID:        1,
				UserID:    1,
				Revoked:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantNewToken: true,
		},
		{
			name: "revoked token",
			storedToken: &domain.RefreshToken{
				Revoked:   true,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantErr: apperrors.ErrUnauthorized,
		},
		{
			name: "expired token",
			storedToken: &domain.RefreshToken{
				Revoked:   false,
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			wantErr: apperrors.ErrUnauthorized,
		},
		{
			name:    "token not found",
			findErr: apperrors.ErrNotFound,
			wantErr: apperrors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := &testutil.MockTokenRepo{
				FindByHashFn: func(_ context.Context, _ string) (*domain.RefreshToken, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return tt.storedToken, nil
				},
				RevokeFn: func(_ context.Context, _ string) error { return nil },
				SaveFn: func(_ context.Context, rt *domain.RefreshToken) error {
					rt.ID = 2
					return nil
				},
			}
			userRepo := &testutil.MockUserRepo{
				FindByIDFn: func(_ context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: id, Role: domain.RoleCustomer}, nil
				},
			}
			authMgr := &testutil.MockAuthManager{
				GenerateAccessTokenFn: func(_ int64, _ domain.Role) (string, error) {
					return "new-access-token", nil
				},
				AccessTTLFn: func() time.Duration { return 15 * time.Minute },
			}

			svc := service.NewTokenService(tokenRepo, userRepo, authMgr, 7*24*time.Hour)
			tokens, err := svc.Refresh(context.Background(), "some-refresh-token")

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "new-access-token", tokens.AccessToken)
			assert.NotEmpty(t, tokens.RefreshToken)
		})
	}
}
