package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(userSvc *MockUserService, tokenSvc *MockTokenService)
		errCode    int
		errMessage string
	}{
		{
			name:    "invalid body",
			body:    `{"email":`,
			errCode: http.StatusBadRequest,
		},
		{
			name: "email conflict",
			body: `{"email":"john@example.com","password":"password123","first_name":"John","last_name":"Doe"}`,
			setup: func(userSvc *MockUserService, _ *MockTokenService) {
				userSvc.RegisterFn = func(_ context.Context, _ *models.Register) (*models.User, error) {
					return nil, apperrors.Conflict("email already taken", nil)
				}
			},
			errCode:    http.StatusConflict,
			errMessage: "email already taken",
		},
		{
			name: "token generation error",
			body: `{"email":"john@example.com","password":"password123","first_name":"John","last_name":"Doe"}`,
			setup: func(userSvc *MockUserService, tokenSvc *MockTokenService) {
				userSvc.RegisterFn = func(_ context.Context, _ *models.Register) (*models.User, error) {
					return &models.User{ID: 1, Role: models.RoleCustomer}, nil
				}
				tokenSvc.GenerateTokenPairFn = func(_ context.Context, _ int64, _ models.Role) (*models.AuthTokens, error) {
					return nil, apperrors.Internal("token generation failed", nil)
				}
			},
			errCode:    http.StatusInternalServerError,
			errMessage: "token generation failed",
		},
		{
			name: "success",
			body: `{"email":"john@example.com","password":"password123","first_name":"John","last_name":"Doe"}`,
			setup: func(userSvc *MockUserService, tokenSvc *MockTokenService) {
				userSvc.RegisterFn = func(_ context.Context, _ *models.Register) (*models.User, error) {
					return &models.User{ID: 1, Email: "john@example.com", Role: models.RoleCustomer}, nil
				}
				tokenSvc.GenerateTokenPairFn = func(_ context.Context, _ int64, _ models.Role) (*models.AuthTokens, error) {
					return &models.AuthTokens{AccessToken: "access", RefreshToken: "refresh"}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &MockUserService{}
			tokenSvc := &MockTokenService{}
			if tt.setup != nil {
				tt.setup(userSvc, tokenSvc)
			}
			h := newTestHandler(&service.Services{User: userSvc, Token: tokenSvc})

			c, w := newTestContext(http.MethodPost, "/auth/register", tt.body)
			h.Register(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusCreated, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(userSvc *MockUserService, tokenSvc *MockTokenService)
		errCode    int
		errMessage string
	}{
		{
			name:    "invalid body",
			body:    `{"email":`,
			errCode: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"john@example.com","password":"wrong"}`,
			setup: func(userSvc *MockUserService, _ *MockTokenService) {
				userSvc.LoginFn = func(_ context.Context, _ *models.Login) (*models.UserRecord, error) {
					return nil, apperrors.BadRequest("invalid credentials", nil)
				}
			},
			errCode:    http.StatusBadRequest,
			errMessage: "invalid credentials",
		},
		{
			name: "success",
			body: `{"email":"john@example.com","password":"password123"}`,
			setup: func(userSvc *MockUserService, tokenSvc *MockTokenService) {
				userSvc.LoginFn = func(_ context.Context, _ *models.Login) (*models.UserRecord, error) {
					return &models.UserRecord{ID: 1, Email: "john@example.com", Role: models.RoleCustomer}, nil
				}
				tokenSvc.GenerateTokenPairFn = func(_ context.Context, _ int64, _ models.Role) (*models.AuthTokens, error) {
					return &models.AuthTokens{AccessToken: "access", RefreshToken: "refresh"}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &MockUserService{}
			tokenSvc := &MockTokenService{}
			if tt.setup != nil {
				tt.setup(userSvc, tokenSvc)
			}
			h := newTestHandler(&service.Services{User: userSvc, Token: tokenSvc})

			c, w := newTestContext(http.MethodPost, "/auth/login", tt.body)
			h.Login(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(tokenSvc *MockTokenService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid body",
			body:       `{}`,
			errCode:    http.StatusBadRequest,
			errMessage: "Key: 'RefreshToken' Error:Field validation for 'RefreshToken' failed on the 'required' tag",
		},
		{
			name: "invalid token",
			body: `{"refresh_token":"invalid"}`,
			setup: func(tokenSvc *MockTokenService) {
				tokenSvc.RefreshFn = func(_ context.Context, _ string) (*models.AuthTokens, error) {
					return nil, apperrors.Unauthorized("invalid or expired token", nil)
				}
			},
			errCode:    http.StatusUnauthorized,
			errMessage: "invalid or expired token",
		},
		{
			name: "success",
			body: `{"refresh_token":"valid-token"}`,
			setup: func(tokenSvc *MockTokenService) {
				tokenSvc.RefreshFn = func(_ context.Context, _ string) (*models.AuthTokens, error) {
					return &models.AuthTokens{AccessToken: "new-access"}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc := &MockTokenService{}
			if tt.setup != nil {
				tt.setup(tokenSvc)
			}
			h := newTestHandler(&service.Services{Token: tokenSvc})

			c, w := newTestContext(http.MethodPost, "/auth/refresh", tt.body)
			h.Refresh(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(tokenSvc *MockTokenService)
		errCode    int
		errMessage string
	}{
		{
			name:    "missing refresh token",
			body:    `{}`,
			errCode: http.StatusBadRequest,
		},
		{
			name: "revoke error",
			body: `{"refresh_token":"some-token"}`,
			setup: func(tokenSvc *MockTokenService) {
				tokenSvc.RevokeFn = func(_ context.Context, _ string) error {
					return apperrors.NotFound("token not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "token not found",
		},
		{
			name: "success",
			body: `{"refresh_token":"valid-token"}`,
			setup: func(tokenSvc *MockTokenService) {
				tokenSvc.RevokeFn = func(_ context.Context, _ string) error { return nil }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc := &MockTokenService{}
			if tt.setup != nil {
				tt.setup(tokenSvc)
			}
			h := newTestHandler(&service.Services{Token: tokenSvc})

			c, w := newTestContext(http.MethodPost, "/auth/logout", tt.body)
			h.Logout(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusNoContent, w.Code)
			assert.Empty(t, w.Body.String())
		})
	}
}
