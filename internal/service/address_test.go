package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

// ── Create ────────────────────────────────────────────────────────────────────

func TestAddressService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int64
		req        *models.CreateAddress
		setup      func(r *MockAddressRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, addr *models.Address)
	}{
		{
			name:   "success — no default",
			userID: 1,
			req: &models.CreateAddress{
				FullName:   "John Doe",
				Phone:      "+1234567890",
				Country:    "Kazakhstan",
				City:       "Almaty",
				Street:     "Abay Ave 1",
				PostalCode: "050000",
			},
			setup: func(r *MockAddressRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.Address")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.Address).ID = 7
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, addr *models.Address) {
				assert.Equal(t, int64(7), addr.ID)
				assert.Equal(t, int64(1), addr.UserID)
				assert.Equal(t, "John Doe", addr.FullName)
				assert.Equal(t, "Almaty", addr.City)
				assert.False(t, addr.CreatedAt.IsZero())
				assert.False(t, addr.UpdatedAt.IsZero())
			},
		},
		{
			name:   "success — with default clears previous defaults",
			userID: 1,
			req:    &models.CreateAddress{FullName: "Jane Doe", City: "Astana", IsDefault: true},
			setup: func(r *MockAddressRepo) {
				r.On("ClearDefault", ctx, int64(1)).Return(nil).Once()
				r.On("Create", ctx, mock.AnythingOfType("*models.Address")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.Address).ID = 8
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, addr *models.Address) {
				assert.True(t, addr.IsDefault)
			},
		},
		{
			name:   "internal error on create",
			userID: 1,
			req:    &models.CreateAddress{FullName: "John"},
			setup: func(r *MockAddressRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.Address")).
					Return(apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAddressRepo)
			if tt.setup != nil {
				tt.setup(repo)
			}

			svc := service.NewAddressService(repo, nil)
			addr, err := svc.Create(ctx, tt.userID, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, addr)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestAddressService_List(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int64
		setup      func(r *MockAddressRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, addrs []*models.Address)
	}{
		{
			name:   "success",
			userID: 1,
			setup: func(r *MockAddressRepo) {
				r.On("ListByUser", ctx, int64(1)).Return([]*models.Address{
					{ID: 1, UserID: 1, City: "Almaty", IsDefault: true},
					{ID: 2, UserID: 1, City: "Astana"},
				}, nil).Once()
			},
			check: func(t *testing.T, addrs []*models.Address) {
				assert.Len(t, addrs, 2)
				assert.Equal(t, "Almaty", addrs[0].City)
			},
		},
		{
			name:   "repo error",
			userID: 1,
			setup: func(r *MockAddressRepo) {
				r.On("ListByUser", ctx, int64(1)).
					Return(nil, apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAddressRepo)
			if tt.setup != nil {
				tt.setup(repo)
			}

			svc := service.NewAddressService(repo, nil)
			addrs, err := svc.List(ctx, tt.userID)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, addrs)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestAddressService_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		addressID  int64
		userID     int64
		req        *models.UpdateAddress
		setup      func(r *MockAddressRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, addr *models.Address)
	}{
		{
			name:      "success",
			addressID: 1,
			userID:    1,
			req:       &models.UpdateAddress{FullName: "Updated Name", City: "New City", Street: "New Street"},
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Address{ID: 1, UserID: 1, City: "Old City"}, nil).Once()
				r.On("Update", ctx, mock.AnythingOfType("*models.Address")).
					Return(nil).Once()
			},
			check: func(t *testing.T, addr *models.Address) {
				assert.Equal(t, "New City", addr.City)
				assert.Equal(t, "Updated Name", addr.FullName)
				assert.False(t, addr.UpdatedAt.IsZero())
			},
		},
		{
			name:      "not owner",
			addressID: 1,
			userID:    99,
			req:       &models.UpdateAddress{},
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Address{ID: 1, UserID: 2}, nil).Once()
			},
			errCode: http.StatusForbidden,
			errMsg:  "cannot update another user's address",
		},
		{
			name:      "not found",
			addressID: 99,
			userID:    1,
			req:       &models.UpdateAddress{},
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("address not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "address not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAddressRepo)
			if tt.setup != nil {
				tt.setup(repo)
			}

			svc := service.NewAddressService(repo, nil)
			addr, err := svc.Update(ctx, tt.addressID, tt.userID, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, addr)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestAddressService_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		addressID  int64
		userID     int64
		setup      func(r *MockAddressRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:      "success",
			addressID: 1,
			userID:    1,
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Address{ID: 1, UserID: 1}, nil).Once()
				r.On("Delete", ctx, int64(1)).Return(nil).Once()
			},
		},
		{
			name:      "not owner",
			addressID: 1,
			userID:    99,
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Address{ID: 1, UserID: 5}, nil).Once()
			},
			errCode: http.StatusForbidden,
			errMsg:  "cannot delete another user's address",
		},
		{
			name:      "not found",
			addressID: 99,
			userID:    1,
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("address not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "address not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAddressRepo)
			if tt.setup != nil {
				tt.setup(repo)
			}

			svc := service.NewAddressService(repo, nil)
			err := svc.Delete(ctx, tt.addressID, tt.userID)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ── SetDefault ────────────────────────────────────────────────────────────────

func TestAddressService_SetDefault(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		addressID  int64
		userID     int64
		setup      func(r *MockAddressRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:      "not owner",
			addressID: 1,
			userID:    99,
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Address{ID: 1, UserID: 5}, nil).Once()
			},
			errCode: http.StatusForbidden,
			errMsg:  "cannot update another user's address",
		},
		{
			name:      "not found",
			addressID: 99,
			userID:    1,
			setup: func(r *MockAddressRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("address not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "address not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAddressRepo)
			if tt.setup != nil {
				tt.setup(repo)
			}

			svc := service.NewAddressService(repo, nil)
			_, err := svc.SetDefault(ctx, tt.addressID, tt.userID)

			require.Error(t, err)
			assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)

			repo.AssertExpectations(t)
		})
	}
}
