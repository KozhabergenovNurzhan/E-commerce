package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestHandler_GetCart(t *testing.T) {
	tests := []struct {
		name       string
		callerID   int64
		setup      func(cartSvc *MockCartService)
		errCode    int
		errMessage string
	}{
		{
			name:     "service error",
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.GetCartFn = func(_ context.Context, _ int64) (*models.Cart, error) {
					return nil, apperrors.Internal("internal server error", nil)
				}
			},
			errCode:    http.StatusInternalServerError,
			errMessage: "internal server error",
		},
		{
			name:     "success — empty cart",
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.GetCartFn = func(_ context.Context, _ int64) (*models.Cart, error) {
					return &models.Cart{Items: []*models.CartItem{}, TotalPrice: 0}, nil
				}
			},
		},
		{
			name:     "success — with items",
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.GetCartFn = func(_ context.Context, userID int64) (*models.Cart, error) {
					return &models.Cart{
						Items:      []*models.CartItem{{ID: 1, Quantity: 2, Subtotal: 20.0}},
						TotalPrice: 20.0,
					}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartSvc := &MockCartService{}
			if tt.setup != nil {
				tt.setup(cartSvc)
			}
			h := newTestHandler(&service.Services{Cart: cartSvc})

			c, w := newTestContext(http.MethodGet, "/cart", "")
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.GetCart(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestHandler_AddToCart(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		callerID   int64
		setup      func(cartSvc *MockCartService)
		errCode    int
		errMessage string
	}{
		{
			name:     "invalid body",
			body:     `{"product_id":`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "product not found",
			body:     `{"product_id":99,"quantity":1}`,
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.AddItemFn = func(_ context.Context, _ int64, _ *models.AddToCart) error {
					return apperrors.NotFound("product not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "product not found",
		},
		{
			name:     "insufficient stock",
			body:     `{"product_id":1,"quantity":100}`,
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.AddItemFn = func(_ context.Context, _ int64, _ *models.AddToCart) error {
					return apperrors.BadRequest("insufficient stock", nil)
				}
			},
			errCode:    http.StatusBadRequest,
			errMessage: "insufficient stock",
		},
		{
			name:     "success",
			body:     `{"product_id":1,"quantity":2}`,
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.AddItemFn = func(_ context.Context, _ int64, _ *models.AddToCart) error { return nil }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartSvc := &MockCartService{}
			if tt.setup != nil {
				tt.setup(cartSvc)
			}
			h := newTestHandler(&service.Services{Cart: cartSvc})

			c, w := newTestContext(http.MethodPost, "/cart/items", tt.body)
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.AddToCart(c)

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

func TestHandler_UpdateCartItem(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		body       string
		callerID   int64
		setup      func(cartSvc *MockCartService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid product id",
			productID:  "xyz",
			body:       `{"quantity":3}`,
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:      "invalid body",
			productID: "1",
			body:      `{"quantity":`,
			callerID:  1,
			errCode:   http.StatusBadRequest,
		},
		{
			name:      "exceeds stock",
			productID: "1",
			body:      `{"quantity":999}`,
			callerID:  1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.UpdateItemFn = func(_ context.Context, _, _ int64, _ *models.UpdateCartItem) error {
					return apperrors.BadRequest("insufficient stock", nil)
				}
			},
			errCode:    http.StatusBadRequest,
			errMessage: "insufficient stock",
		},
		{
			name:      "success",
			productID: "1",
			body:      `{"quantity":3}`,
			callerID:  1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.UpdateItemFn = func(_ context.Context, userID, productID int64, req *models.UpdateCartItem) error {
					assert.Equal(t, int64(1), userID)
					assert.Equal(t, int64(1), productID)
					assert.Equal(t, 3, req.Quantity)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartSvc := &MockCartService{}
			if tt.setup != nil {
				tt.setup(cartSvc)
			}
			h := newTestHandler(&service.Services{Cart: cartSvc})

			c, w := newTestContext(http.MethodPut, "/cart/items/"+tt.productID, tt.body)
			c.Params = gin.Params{{Key: "productId", Value: tt.productID}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.UpdateCartItem(c)

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

func TestHandler_RemoveFromCart(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		callerID   int64
		setup      func(cartSvc *MockCartService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid product id",
			productID:  "xyz",
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:      "service error",
			productID: "99",
			callerID:  1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.RemoveItemFn = func(_ context.Context, _, _ int64) error {
					return apperrors.Internal("internal server error", nil)
				}
			},
			errCode:    http.StatusInternalServerError,
			errMessage: "internal server error",
		},
		{
			name:      "success",
			productID: "1",
			callerID:  1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.RemoveItemFn = func(_ context.Context, userID, productID int64) error {
					assert.Equal(t, int64(1), userID)
					assert.Equal(t, int64(1), productID)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartSvc := &MockCartService{}
			if tt.setup != nil {
				tt.setup(cartSvc)
			}
			h := newTestHandler(&service.Services{Cart: cartSvc})

			c, w := newTestContext(http.MethodDelete, "/cart/items/"+tt.productID, "")
			c.Params = gin.Params{{Key: "productId", Value: tt.productID}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.RemoveFromCart(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusNoContent, w.Code)
			assert.Empty(t, w.Body.String())
		})
	}
}

func TestHandler_ClearCart(t *testing.T) {
	tests := []struct {
		name       string
		callerID   int64
		setup      func(cartSvc *MockCartService)
		errCode    int
		errMessage string
	}{
		{
			name:     "service error",
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.ClearFn = func(_ context.Context, _ int64) error {
					return apperrors.Internal("internal server error", nil)
				}
			},
			errCode:    http.StatusInternalServerError,
			errMessage: "internal server error",
		},
		{
			name:     "success",
			callerID: 1,
			setup: func(cartSvc *MockCartService) {
				cartSvc.ClearFn = func(_ context.Context, userID int64) error {
					assert.Equal(t, int64(1), userID)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartSvc := &MockCartService{}
			if tt.setup != nil {
				tt.setup(cartSvc)
			}
			h := newTestHandler(&service.Services{Cart: cartSvc})

			c, w := newTestContext(http.MethodDelete, "/cart", "")
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.ClearCart(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusNoContent, w.Code)
			assert.Empty(t, w.Body.String())
		})
	}
}
