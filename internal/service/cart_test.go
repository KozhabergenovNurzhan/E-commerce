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

func TestCartService_AddItem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		req        *models.AddToCart
		setup      func(cartRepo *MockCartRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name: "success",
			req:  &models.AddToCart{ProductID: 1, Quantity: 2},
			setup: func(cartRepo *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Stock: 10}, nil).Once()
				cartRepo.On("Upsert", ctx, mock.AnythingOfType("*models.CartItemRecord")).
					Return(nil).Once()
			},
		},
		{
			name: "product not found",
			req:  &models.AddToCart{ProductID: 99, Quantity: 1},
			setup: func(_ *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
		{
			name: "insufficient stock",
			req:  &models.AddToCart{ProductID: 1, Quantity: 5},
			setup: func(_ *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Stock: 3}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "insufficient stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := new(MockCartRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(cartRepo, productRepo)
			}

			svc := service.NewCartService(cartRepo, productRepo)
			err := svc.AddItem(ctx, 1, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			cartRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_GetCart(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(cartRepo *MockCartRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, cart *models.Cart)
	}{
		{
			name: "empty cart",
			setup: func(cartRepo *MockCartRepo, _ *MockProductRepo) {
				cartRepo.On("FindByUserID", ctx, int64(1)).
					Return([]*models.CartItemRecord{}, nil).Once()
			},
			check: func(t *testing.T, cart *models.Cart) {
				assert.Empty(t, cart.Items)
				assert.Equal(t, 0.0, cart.TotalPrice)
			},
		},
		{
			name: "calculates subtotals and total",
			setup: func(cartRepo *MockCartRepo, productRepo *MockProductRepo) {
				cartRepo.On("FindByUserID", ctx, int64(1)).
					Return([]*models.CartItemRecord{
						{ID: 1, UserID: 1, ProductID: 10, Quantity: 2},
						{ID: 2, UserID: 1, ProductID: 20, Quantity: 3},
					}, nil).Once()
				productRepo.On("FindByID", ctx, int64(10)).
					Return(&models.Product{ID: 10, Price: 5.0}, nil).Once()
				productRepo.On("FindByID", ctx, int64(20)).
					Return(&models.Product{ID: 20, Price: 10.0}, nil).Once()
			},
			check: func(t *testing.T, cart *models.Cart) {
				require.Len(t, cart.Items, 2)
				assert.Equal(t, 40.0, cart.TotalPrice)
				assert.Equal(t, 10.0, cart.Items[0].Subtotal)
				assert.Equal(t, 30.0, cart.Items[1].Subtotal)
			},
		},
		{
			name: "skips items whose product is no longer available",
			setup: func(cartRepo *MockCartRepo, productRepo *MockProductRepo) {
				cartRepo.On("FindByUserID", ctx, int64(1)).
					Return([]*models.CartItemRecord{
						{ID: 1, UserID: 1, ProductID: 10, Quantity: 1},
						{ID: 2, UserID: 1, ProductID: 99, Quantity: 1},
					}, nil).Once()
				productRepo.On("FindByID", ctx, int64(10)).
					Return(&models.Product{ID: 10, Price: 20.0}, nil).Once()
				productRepo.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			check: func(t *testing.T, cart *models.Cart) {
				require.Len(t, cart.Items, 1)
				assert.Equal(t, 20.0, cart.TotalPrice)
			},
		},
		{
			name: "repo error",
			setup: func(cartRepo *MockCartRepo, _ *MockProductRepo) {
				cartRepo.On("FindByUserID", ctx, int64(1)).
					Return(nil, apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := new(MockCartRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(cartRepo, productRepo)
			}

			svc := service.NewCartService(cartRepo, productRepo)
			cart, err := svc.GetCart(ctx, 1)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, cart)
			}

			cartRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_UpdateItem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		quantity   int
		setup      func(cartRepo *MockCartRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:     "success",
			quantity: 2,
			setup: func(cartRepo *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Stock: 10}, nil).Once()
				cartRepo.On("Upsert", ctx, mock.AnythingOfType("*models.CartItemRecord")).
					Return(nil).Once()
			},
		},
		{
			name:     "product not found",
			quantity: 1,
			setup: func(_ *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
		{
			name:     "exceeds stock",
			quantity: 99,
			setup: func(_ *MockCartRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Stock: 5}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "insufficient stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := new(MockCartRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(cartRepo, productRepo)
			}

			svc := service.NewCartService(cartRepo, productRepo)
			err := svc.UpdateItem(ctx, 1, 1, &models.UpdateCartItem{Quantity: tt.quantity})

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			cartRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int64
		productID  int64
		setup      func(r *MockCartRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:      "success",
			userID:    7,
			productID: 3,
			setup: func(r *MockCartRepo) {
				r.On("Delete", ctx, int64(7), int64(3)).Return(nil).Once()
			},
		},
		{
			name:      "repo error",
			userID:    1,
			productID: 1,
			setup: func(r *MockCartRepo) {
				r.On("Delete", ctx, int64(1), int64(1)).
					Return(apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := new(MockCartRepo)
			if tt.setup != nil {
				tt.setup(cartRepo)
			}

			svc := service.NewCartService(cartRepo, nil)
			err := svc.RemoveItem(ctx, tt.userID, tt.productID)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			cartRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_Clear(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int64
		setup      func(r *MockCartRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:   "success",
			userID: 5,
			setup: func(r *MockCartRepo) {
				r.On("Clear", ctx, int64(5)).Return(nil).Once()
			},
		},
		{
			name:   "repo error",
			userID: 5,
			setup: func(r *MockCartRepo) {
				r.On("Clear", ctx, int64(5)).
					Return(apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartRepo := new(MockCartRepo)
			if tt.setup != nil {
				tt.setup(cartRepo)
			}

			svc := service.NewCartService(cartRepo, nil)
			err := svc.Clear(ctx, tt.userID)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			cartRepo.AssertExpectations(t)
		})
	}
}
