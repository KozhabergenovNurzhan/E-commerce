package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
)

func TestGenerateTokenPair(t *testing.T) {
	tests := []struct {
		name          string
		generateFn    func(int64, models.Role) (string, error)
		saveFn        func(context.Context, *models.RefreshToken) error
		wantCode      int
		wantAccessTok string
	}{
		{
			name: "success",
			generateFn: func(_ int64, _ models.Role) (string, error) {
				return "access-token", nil
			},
			saveFn: func(_ context.Context, rt *models.RefreshToken) error {
				rt.ID = 1
				return nil
			},
			wantAccessTok: "access-token",
		},
		{
			name: "auth manager failure",
			generateFn: func(_ int64, _ models.Role) (string, error) {
				return "", apperrors.Internal("token generation failed", nil)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMgr := &testutil.MockAuthManager{
				GenerateAccessTokenFn: tt.generateFn,
				AccessTTLFn:           func() time.Duration { return 15 * time.Minute },
			}
			tokenRepo := &testutil.MockTokenRepo{SaveFn: tt.saveFn}

			svc := service.NewTokenService(nil, tokenRepo, nil, authMgr, 7*24*time.Hour)
			tokens, err := svc.GenerateTokenPair(context.Background(), 1, models.RoleCustomer)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
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
		storedToken  *models.RefreshToken
		findErr      error
		wantCode     int
		wantNewToken bool
	}{
		{
			name: "success",
			storedToken: &models.RefreshToken{
				ID:        1,
				UserID:    1,
				Revoked:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantNewToken: true,
		},
		{
			name: "revoked token",
			storedToken: &models.RefreshToken{
				Revoked:   true,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			storedToken: &models.RefreshToken{
				Revoked:   false,
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "token not found",
			findErr:  apperrors.NotFound("token not found", nil),
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := &testutil.MockTokenRepo{
				FindByHashFn: func(_ context.Context, _ string) (*models.RefreshToken, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return tt.storedToken, nil
				},
				RevokeFn: func(_ context.Context, _ string) error { return nil },
				SaveFn: func(_ context.Context, rt *models.RefreshToken) error {
					rt.ID = 2
					return nil
				},
			}
			userRepo := &testutil.MockUserRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.UserRecord, error) {
					return &models.UserRecord{ID: id, Role: models.RoleCustomer}, nil
				},
			}
			authMgr := &testutil.MockAuthManager{
				GenerateAccessTokenFn: func(_ int64, _ models.Role) (string, error) {
					return "new-access-token", nil
				},
				AccessTTLFn: func() time.Duration { return 15 * time.Minute },
			}

			svc := service.NewTokenService(nil, tokenRepo, userRepo, authMgr, 7*24*time.Hour)
			tokens, err := svc.Refresh(context.Background(), "some-refresh-token")

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "new-access-token", tokens.AccessToken)
			assert.NotEmpty(t, tokens.RefreshToken)
		})
	}
}
