package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		req        *models.Register
		setup      func(r *MockUserRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, u *models.User)
	}{
		{
			name: "success",
			req: &models.Register{
				Email:     "john@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setup: func(r *MockUserRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.UserRecord")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.UserRecord).ID = 1
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, u *models.User) {
				assert.Equal(t, int64(1), u.ID)
				assert.Equal(t, "john@example.com", u.Email)
				assert.Equal(t, models.RoleCustomer, u.Role)
			},
		},
		{
			name: "email already taken",
			req: &models.Register{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func(r *MockUserRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.UserRecord")).
					Return(apperrors.Conflict("email already taken", nil)).Once()
			},
			errCode: http.StatusConflict,
			errMsg:  "email already taken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepo)
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := service.NewUserService(userRepo)
			user, err := svc.Register(ctx, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, user)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	tests := []struct {
		name       string
		req        *models.Login
		setup      func(r *MockUserRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, u *models.UserRecord)
	}{
		{
			name: "success",
			req:  &models.Login{Email: "john@example.com", Password: "password123"},
			setup: func(r *MockUserRepo) {
				r.On("FindByEmail", ctx, "john@example.com").
					Return(&models.UserRecord{ID: 1, Email: "john@example.com", PasswordHash: string(hash), Role: models.RoleCustomer}, nil).Once()
			},
			check: func(t *testing.T, u *models.UserRecord) {
				assert.Equal(t, int64(1), u.ID)
			},
		},
		{
			name: "wrong password",
			req:  &models.Login{Email: "john@example.com", Password: "wrongpass"},
			setup: func(r *MockUserRepo) {
				r.On("FindByEmail", ctx, "john@example.com").
					Return(&models.UserRecord{PasswordHash: string(hash)}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "invalid credentials",
		},
		{
			name: "user not found returns bad request to avoid enumeration",
			req:  &models.Login{Email: "nobody@example.com", Password: "password123"},
			setup: func(r *MockUserRepo) {
				r.On("FindByEmail", ctx, "nobody@example.com").
					Return(nil, apperrors.NotFound("user not found", nil)).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepo)
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := service.NewUserService(userRepo)
			user, err := svc.Login(ctx, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, user)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
