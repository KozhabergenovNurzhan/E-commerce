package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
)

func assertCode(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected *apperrors.AppError, got %T: %v", err, err)
	assert.Equal(t, code, appErr.Code)
}

func TestCartAddItem(t *testing.T) {
	tests := []struct {
		name        string
		req         *models.AddToCart
		findProduct func(ctx context.Context, id int64) (*models.Product, error)
		upsert      func(ctx context.Context, item *models.CartItemRecord) error
		wantCode    int
	}{
		{
			name: "success",
			req:  &models.AddToCart{ProductID: 1, Quantity: 2},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Stock: 10}, nil
			},
			upsert: func(_ context.Context, item *models.CartItemRecord) error { return nil },
		},
		{
			name: "product not found",
			req:  &models.AddToCart{ProductID: 99, Quantity: 1},
			findProduct: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "insufficient stock",
			req:  &models.AddToCart{ProductID: 1, Quantity: 5},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Stock: 3}, nil
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{UpsertFn: tt.upsert}
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}

			err := service.NewCartService(cartRepo, productRepo).
				AddItem(context.Background(), 1, tt.req)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCartGetCart(t *testing.T) {
	tests := []struct {
		name        string
		items       []*models.CartItemRecord
		findProduct func(ctx context.Context, id int64) (*models.Product, error)
		check       func(t *testing.T, resp *models.Cart)
	}{
		{
			name:  "empty cart",
			items: []*models.CartItemRecord{},
			findProduct: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			check: func(t *testing.T, resp *models.Cart) {
				assert.Empty(t, resp.Items)
				assert.Equal(t, 0.0, resp.TotalPrice)
			},
		},
		{
			name: "calculates subtotals and total",
			items: []*models.CartItemRecord{
				{ID: 1, UserID: 1, ProductID: 10, Quantity: 2},
				{ID: 2, UserID: 1, ProductID: 20, Quantity: 3},
			},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				prices := map[int64]float64{10: 5.0, 20: 10.0}
				return &models.Product{ID: id, Price: prices[id]}, nil
			},
			check: func(t *testing.T, resp *models.Cart) {
				require.Len(t, resp.Items, 2)
				assert.Equal(t, 40.0, resp.TotalPrice)
				assert.Equal(t, 10.0, resp.Items[0].Subtotal)
				assert.Equal(t, 30.0, resp.Items[1].Subtotal)
			},
		},
		{
			name: "skips items whose product is no longer available",
			items: []*models.CartItemRecord{
				{ID: 1, UserID: 1, ProductID: 10, Quantity: 1},
				{ID: 2, UserID: 1, ProductID: 99, Quantity: 1},
			},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				if id == 10 {
					return &models.Product{ID: id, Price: 20.0}, nil
				}
				return nil, apperrors.NotFound("product not found", nil)
			},
			check: func(t *testing.T, resp *models.Cart) {
				require.Len(t, resp.Items, 1)
				assert.Equal(t, 20.0, resp.TotalPrice)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{
				FindByUserIDFn: func(_ context.Context, _ int64) ([]*models.CartItemRecord, error) {
					return tt.items, nil
				},
			}
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}

			resp, err := service.NewCartService(cartRepo, productRepo).
				GetCart(context.Background(), 1)

			require.NoError(t, err)
			tt.check(t, resp)
		})
	}
}

func TestCartGetCart_RepoError(t *testing.T) {
	cartRepo := &testutil.MockCartRepo{
		FindByUserIDFn: func(_ context.Context, _ int64) ([]*models.CartItemRecord, error) {
			return nil, apperrors.Internal("internal server error", nil)
		},
	}

	_, err := service.NewCartService(cartRepo, nil).GetCart(context.Background(), 1)
	assertCode(t, err, http.StatusInternalServerError)
}

func TestCartUpdateItem(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		findProduct func(ctx context.Context, id int64) (*models.Product, error)
		wantCode    int
	}{
		{
			name:     "success",
			quantity: 2,
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Stock: 10}, nil
			},
		},
		{
			name:     "product not found",
			quantity: 1,
			findProduct: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "exceeds stock",
			quantity: 99,
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Stock: 5}, nil
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{
				UpsertFn: func(_ context.Context, _ *models.CartItemRecord) error { return nil },
			}
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}

			err := service.NewCartService(cartRepo, productRepo).
				UpdateItem(context.Background(), 1, 1, &models.UpdateCartItem{Quantity: tt.quantity})

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCartRemoveItem(t *testing.T) {
	called := false
	cartRepo := &testutil.MockCartRepo{
		DeleteFn: func(_ context.Context, userID, productID int64) error {
			called = true
			assert.Equal(t, int64(7), userID)
			assert.Equal(t, int64(3), productID)
			return nil
		},
	}

	err := service.NewCartService(cartRepo, nil).RemoveItem(context.Background(), 7, 3)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestCartClear(t *testing.T) {
	called := false
	cartRepo := &testutil.MockCartRepo{
		ClearFn: func(_ context.Context, userID int64) error {
			called = true
			assert.Equal(t, int64(5), userID)
			return nil
		},
	}

	err := service.NewCartService(cartRepo, nil).Clear(context.Background(), 5)
	require.NoError(t, err)
	assert.True(t, called)
}
