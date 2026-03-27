package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

func TestCartAddItem(t *testing.T) {
	tests := []struct {
		name        string
		req         *domain.AddToCartRequest
		findProduct func(ctx context.Context, id int64) (*domain.Product, error)
		upsert      func(ctx context.Context, item *domain.CartItem) error
		wantErr     error
	}{
		{
			name: "success",
			req:  &domain.AddToCartRequest{ProductID: 1, Quantity: 2},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Stock: 10}, nil
			},
			upsert: func(_ context.Context, item *domain.CartItem) error { return nil },
		},
		{
			name: "product not found",
			req:  &domain.AddToCartRequest{ProductID: 99, Quantity: 1},
			findProduct: func(_ context.Context, _ int64) (*domain.Product, error) {
				return nil, apperrors.ErrNotFound
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "insufficient stock",
			req:  &domain.AddToCartRequest{ProductID: 1, Quantity: 5},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Stock: 3}, nil
			},
			wantErr: apperrors.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{UpsertFn: tt.upsert}
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}

			err := service.NewCartService(cartRepo, productRepo).
				AddItem(context.Background(), 1, tt.req)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestCartGetCart(t *testing.T) {
	tests := []struct {
		name        string
		items       []*domain.CartItem
		findProduct func(ctx context.Context, id int64) (*domain.Product, error)
		check       func(t *testing.T, resp *domain.CartResponse)
	}{
		{
			name:  "empty cart",
			items: []*domain.CartItem{},
			findProduct: func(_ context.Context, _ int64) (*domain.Product, error) {
				return nil, apperrors.ErrNotFound
			},
			check: func(t *testing.T, resp *domain.CartResponse) {
				assert.Empty(t, resp.Items)
				assert.Equal(t, 0.0, resp.TotalPrice)
			},
		},
		{
			name: "calculates subtotals and total",
			items: []*domain.CartItem{
				{ID: 1, UserID: 1, ProductID: 10, Quantity: 2},
				{ID: 2, UserID: 1, ProductID: 20, Quantity: 3},
			},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				prices := map[int64]float64{10: 5.0, 20: 10.0}
				return &domain.Product{ID: id, Price: prices[id]}, nil
			},
			check: func(t *testing.T, resp *domain.CartResponse) {
				require.Len(t, resp.Items, 2)
				// item 1: 2 * 5.0 = 10; item 2: 3 * 10.0 = 30; total = 40
				assert.Equal(t, 40.0, resp.TotalPrice)
				assert.Equal(t, 10.0, resp.Items[0].Subtotal)
				assert.Equal(t, 30.0, resp.Items[1].Subtotal)
			},
		},
		{
			name: "skips items whose product is no longer available",
			items: []*domain.CartItem{
				{ID: 1, UserID: 1, ProductID: 10, Quantity: 1},
				{ID: 2, UserID: 1, ProductID: 99, Quantity: 1},
			},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				if id == 10 {
					return &domain.Product{ID: id, Price: 20.0}, nil
				}
				return nil, apperrors.ErrNotFound
			},
			check: func(t *testing.T, resp *domain.CartResponse) {
				require.Len(t, resp.Items, 1)
				assert.Equal(t, 20.0, resp.TotalPrice)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{
				FindByUserIDFn: func(_ context.Context, _ int64) ([]*domain.CartItem, error) {
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
	repoErr := errors.New("db down")
	cartRepo := &testutil.MockCartRepo{
		FindByUserIDFn: func(_ context.Context, _ int64) ([]*domain.CartItem, error) {
			return nil, repoErr
		},
	}

	_, err := service.NewCartService(cartRepo, nil).GetCart(context.Background(), 1)
	assert.Equal(t, repoErr, err)
}

func TestCartUpdateItem(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		findProduct func(ctx context.Context, id int64) (*domain.Product, error)
		wantErr     error
	}{
		{
			name:     "success",
			quantity: 2,
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Stock: 10}, nil
			},
		},
		{
			name:     "product not found",
			quantity: 1,
			findProduct: func(_ context.Context, _ int64) (*domain.Product, error) {
				return nil, apperrors.ErrNotFound
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:     "exceeds stock",
			quantity: 99,
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Stock: 5}, nil
			},
			wantErr: apperrors.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := &testutil.MockCartRepo{
				UpsertFn: func(_ context.Context, _ *domain.CartItem) error { return nil },
			}
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}

			err := service.NewCartService(cartRepo, productRepo).
				UpdateItem(context.Background(), 1, 1, &domain.UpdateCartItemRequest{Quantity: tt.quantity})

			assert.Equal(t, tt.wantErr, err)
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
