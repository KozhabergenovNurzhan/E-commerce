package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestTokenService_GenerateTokenPair(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(tokenRepo *MockTokenRepo, authMgr *MockAuthManager)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, tokens *models.AuthTokens)
	}{
		{
			name: "success",
			setup: func(tokenRepo *MockTokenRepo, authMgr *MockAuthManager) {
				authMgr.On("GenerateAccessToken", int64(1), models.RoleCustomer).
					Return("access-token", nil).Once()
				authMgr.On("AccessTTL").Return(15 * time.Minute).Once()
				tokenRepo.On("Save", ctx, mock.AnythingOfType("*models.RefreshToken")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.RefreshToken).ID = 1
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, tokens *models.AuthTokens) {
				assert.Equal(t, "access-token", tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Equal(t, int64(900), tokens.ExpiresIn)
			},
		},
		{
			name: "auth manager failure",
			setup: func(_ *MockTokenRepo, authMgr *MockAuthManager) {
				authMgr.On("GenerateAccessToken", int64(1), models.RoleCustomer).
					Return("", apperrors.Internal("token generation failed", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "token generation failed",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := new(MockTokenRepo)
			authMgr := new(MockAuthManager)
			if tt.setup != nil {
				tt.setup(tokenRepo, authMgr)
			}

			svc := service.NewTokenService(nil, tokenRepo, nil, authMgr, 7*24*time.Hour)
			tokens, err := svc.GenerateTokenPair(ctx, 1, models.RoleCustomer)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, tokens)
			}

			tokenRepo.AssertExpectations(t)
			authMgr.AssertExpectations(t)
		})
	}
}

func TestTokenService_Refresh(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(tokenRepo *MockTokenRepo, userRepo *MockUserRepo, authMgr *MockAuthManager)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, tokens *models.AuthTokens)
	}{
		{
			name: "success",
			setup: func(tokenRepo *MockTokenRepo, userRepo *MockUserRepo, authMgr *MockAuthManager) {
				tokenRepo.On("FindByHash", ctx, mock.AnythingOfType("string")).
					Return(&models.RefreshToken{
						ID:        1,
						UserID:    1,
						Revoked:   false,
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil).Once()
				userRepo.On("FindByID", ctx, int64(1)).
					Return(&models.UserRecord{ID: 1, Role: models.RoleCustomer}, nil).Once()
				tokenRepo.On("Revoke", ctx, mock.AnythingOfType("string")).
					Return(nil).Once()
				authMgr.On("GenerateAccessToken", int64(1), models.RoleCustomer).
					Return("new-access-token", nil).Once()
				authMgr.On("AccessTTL").Return(15 * time.Minute).Once()
				tokenRepo.On("Save", ctx, mock.AnythingOfType("*models.RefreshToken")).
					Return(nil).Once()
			},
			check: func(t *testing.T, tokens *models.AuthTokens) {
				assert.Equal(t, "new-access-token", tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
			},
		},
		{
			name: "revoked token",
			setup: func(tokenRepo *MockTokenRepo, _ *MockUserRepo, _ *MockAuthManager) {
				tokenRepo.On("FindByHash", ctx, mock.AnythingOfType("string")).
					Return(&models.RefreshToken{
						Revoked:   true,
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil).Once()
			},
			errCode: http.StatusUnauthorized,
			errMsg:  "invalid or expired token",
		},
		{
			name: "expired token",
			setup: func(tokenRepo *MockTokenRepo, _ *MockUserRepo, _ *MockAuthManager) {
				tokenRepo.On("FindByHash", ctx, mock.AnythingOfType("string")).
					Return(&models.RefreshToken{
						Revoked:   false,
						ExpiresAt: time.Now().Add(-time.Hour),
					}, nil).Once()
			},
			errCode: http.StatusUnauthorized,
			errMsg:  "invalid or expired token",
		},
		{
			name: "token not found",
			setup: func(tokenRepo *MockTokenRepo, _ *MockUserRepo, _ *MockAuthManager) {
				tokenRepo.On("FindByHash", ctx, mock.AnythingOfType("string")).
					Return(nil, apperrors.NotFound("token not found", nil)).Once()
			},
			errCode: http.StatusUnauthorized,
			errMsg:  "invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := new(MockTokenRepo)
			userRepo := new(MockUserRepo)
			authMgr := new(MockAuthManager)
			if tt.setup != nil {
				tt.setup(tokenRepo, userRepo, authMgr)
			}

			svc := service.NewTokenService(nil, tokenRepo, userRepo, authMgr, 7*24*time.Hour)
			tokens, err := svc.Refresh(ctx, "some-refresh-token")

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, tokens)
			}

			tokenRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			authMgr.AssertExpectations(t)
		})
	}
}
