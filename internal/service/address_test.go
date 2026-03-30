package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAddressService(repo *testutil.MockAddressRepo) *service.AddressService {
	return service.NewAddressService(repo, nil)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestAddressCreate(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		CreateFn: func(_ context.Context, a *models.Address) error {
			a.ID = 7
			return nil
		},
	}

	req := &models.CreateAddress{
		FullName:   "John Doe",
		Phone:      "+1234567890",
		Country:    "Kazakhstan",
		City:       "Almaty",
		Street:     "Abay Ave 1",
		PostalCode: "050000",
	}

	addr, err := newAddressService(repo).Create(context.Background(), 1, req)

	require.NoError(t, err)
	assert.Equal(t, int64(7), addr.ID)
	assert.Equal(t, int64(1), addr.UserID)
	assert.Equal(t, "John Doe", addr.FullName)
	assert.Equal(t, "Almaty", addr.City)
	assert.False(t, addr.CreatedAt.IsZero())
	assert.False(t, addr.UpdatedAt.IsZero())
}

func TestAddressCreate_WithDefault(t *testing.T) {
	clearDefaultCalled := false
	repo := &testutil.MockAddressRepo{
		ClearDefaultFn: func(_ context.Context, userID int64) error {
			clearDefaultCalled = true
			assert.Equal(t, int64(1), userID)
			return nil
		},
		CreateFn: func(_ context.Context, a *models.Address) error {
			a.ID = 8
			return nil
		},
	}

	req := &models.CreateAddress{
		FullName:  "Jane Doe",
		City:      "Astana",
		IsDefault: true,
	}

	addr, err := newAddressService(repo).Create(context.Background(), 1, req)

	require.NoError(t, err)
	assert.True(t, clearDefaultCalled, "ClearDefault must be called when IsDefault=true")
	assert.True(t, addr.IsDefault)
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestAddressList(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		ListByUserFn: func(_ context.Context, userID int64) ([]*models.Address, error) {
			return []*models.Address{
				{ID: 1, UserID: userID, City: "Almaty", IsDefault: true},
				{ID: 2, UserID: userID, City: "Astana"},
			}, nil
		},
	}

	addrs, err := newAddressService(repo).List(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, addrs, 2)
	assert.Equal(t, "Almaty", addrs[0].City)
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestAddressUpdate(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, id int64) (*models.Address, error) {
			return &models.Address{ID: id, UserID: 1, City: "Old City"}, nil
		},
		UpdateFn: func(_ context.Context, _ *models.Address) error {
			return nil
		},
	}

	req := &models.UpdateAddress{
		FullName: "Updated Name",
		City:     "New City",
		Street:   "New Street",
	}

	addr, err := newAddressService(repo).Update(context.Background(), 1, 1, req)

	require.NoError(t, err)
	assert.Equal(t, "New City", addr.City)
	assert.Equal(t, "Updated Name", addr.FullName)
	assert.False(t, addr.UpdatedAt.IsZero())
}

func TestAddressUpdate_NotOwner(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, id int64) (*models.Address, error) {
			return &models.Address{ID: id, UserID: 2}, nil // owner is userID=2
		},
	}

	_, err := newAddressService(repo).Update(context.Background(), 1, 99, &models.UpdateAddress{}) // caller is 99

	assertCode(t, err, http.StatusForbidden)
}

func TestAddressUpdate_NotFound(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, _ int64) (*models.Address, error) {
			return nil, apperrors.NotFound("address not found", nil)
		},
	}

	_, err := newAddressService(repo).Update(context.Background(), 99, 1, &models.UpdateAddress{})

	assertCode(t, err, http.StatusNotFound)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestAddressDelete(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, id int64) (*models.Address, error) {
			return &models.Address{ID: id, UserID: 1}, nil
		},
		DeleteFn: func(_ context.Context, _ int64) error {
			return nil
		},
	}

	err := newAddressService(repo).Delete(context.Background(), 1, 1)

	require.NoError(t, err)
}

func TestAddressDelete_NotOwner(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, id int64) (*models.Address, error) {
			return &models.Address{ID: id, UserID: 5}, nil
		},
	}

	err := newAddressService(repo).Delete(context.Background(), 1, 99)

	assertCode(t, err, http.StatusForbidden)
}

func TestAddressDelete_NotFound(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, _ int64) (*models.Address, error) {
			return nil, apperrors.NotFound("address not found", nil)
		},
	}

	err := newAddressService(repo).Delete(context.Background(), 99, 1)

	assertCode(t, err, http.StatusNotFound)
}

// ── SetDefault ────────────────────────────────────────────────────────────────

func TestAddressSetDefault_NotOwner(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, id int64) (*models.Address, error) {
			return &models.Address{ID: id, UserID: 5}, nil // owner is 5
		},
	}

	_, err := newAddressService(repo).SetDefault(context.Background(), 1, 99) // caller is 99

	assertCode(t, err, http.StatusForbidden)
}

func TestAddressSetDefault_NotFound(t *testing.T) {
	repo := &testutil.MockAddressRepo{
		FindByIDFn: func(_ context.Context, _ int64) (*models.Address, error) {
			return nil, apperrors.NotFound("address not found", nil)
		},
	}

	_, err := newAddressService(repo).SetDefault(context.Background(), 99, 1)

	assertCode(t, err, http.StatusNotFound)
}
