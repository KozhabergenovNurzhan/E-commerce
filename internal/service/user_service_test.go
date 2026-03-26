package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.RegisterRequest
		createFn func(ctx context.Context, user *domain.User) error
		wantErr error
		check   func(t *testing.T, resp *domain.UserResponse)
	}{
		{
			name: "success",
			req: &domain.RegisterRequest{
				Email:     "john@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			createFn: func(_ context.Context, user *domain.User) error {
				user.ID = 1
				return nil
			},
			check: func(t *testing.T, resp *domain.UserResponse) {
				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "john@example.com", resp.Email)
				assert.Equal(t, domain.RoleCustomer, resp.Role)
			},
		},
		{
			name: "email already taken",
			req: &domain.RegisterRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			createFn: func(_ context.Context, _ *domain.User) error {
				return apperrors.ErrConflict
			},
			wantErr: apperrors.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{CreateFn: tt.createFn}
			resp, err := service.NewUserService(repo).Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, resp)
		})
	}
}

func TestLogin(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		req           *domain.LoginRequest
		findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
		wantErr       error
		wantUserID    int64
	}{
		{
			name: "success",
			req:  &domain.LoginRequest{Email: "john@example.com", Password: "password123"},
			findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: 1, Email: email, PasswordHash: string(hash), Role: domain.RoleCustomer}, nil
			},
			wantUserID: 1,
		},
		{
			name: "wrong password",
			req:  &domain.LoginRequest{Email: "john@example.com", Password: "wrongpass"},
			findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: 1, Email: email, PasswordHash: string(hash)}, nil
			},
			wantErr: apperrors.ErrBadRequest,
		},
		{
			name: "user not found returns bad request to avoid enumeration",
			req:  &domain.LoginRequest{Email: "nobody@example.com", Password: "password123"},
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, apperrors.ErrNotFound
			},
			wantErr: apperrors.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByEmailFn: tt.findByEmailFn}
			user, err := service.NewUserService(repo).Login(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, user.ID)
		})
	}
}
