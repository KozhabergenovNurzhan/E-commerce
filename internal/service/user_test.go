package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name     string
		req      *models.Register
		createFn func(ctx context.Context, user *models.UserRecord) error
		wantCode int
		check    func(t *testing.T, resp *models.User)
	}{
		{
			name: "success",
			req: &models.Register{
				Email:     "john@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			createFn: func(_ context.Context, user *models.UserRecord) error {
				user.ID = 1
				return nil
			},
			check: func(t *testing.T, resp *models.User) {
				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "john@example.com", resp.Email)
				assert.Equal(t, models.RoleCustomer, resp.Role)
			},
		},
		{
			name: "email already taken",
			req: &models.Register{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			createFn: func(_ context.Context, _ *models.UserRecord) error {
				return apperrors.Conflict("email already taken", nil)
			},
			wantCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{CreateFn: tt.createFn}
			resp, err := service.NewUserService(repo).Register(context.Background(), tt.req)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
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
		req           *models.Login
		findByEmailFn func(ctx context.Context, email string) (*models.UserRecord, error)
		wantCode      int
		wantUserID    int64
	}{
		{
			name: "success",
			req:  &models.Login{Email: "john@example.com", Password: "password123"},
			findByEmailFn: func(_ context.Context, email string) (*models.UserRecord, error) {
				return &models.UserRecord{ID: 1, Email: email, PasswordHash: string(hash), Role: models.RoleCustomer}, nil
			},
			wantUserID: 1,
		},
		{
			name: "wrong password",
			req:  &models.Login{Email: "john@example.com", Password: "wrongpass"},
			findByEmailFn: func(_ context.Context, email string) (*models.UserRecord, error) {
				return &models.UserRecord{ID: 1, Email: email, PasswordHash: string(hash)}, nil
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "user not found returns bad request to avoid enumeration",
			req:  &models.Login{Email: "nobody@example.com", Password: "password123"},
			findByEmailFn: func(_ context.Context, _ string) (*models.UserRecord, error) {
				return nil, apperrors.NotFound("user not found", nil)
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByEmailFn: tt.findByEmailFn}
			user, err := service.NewUserService(repo).Login(context.Background(), tt.req)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, user.ID)
		})
	}
}
