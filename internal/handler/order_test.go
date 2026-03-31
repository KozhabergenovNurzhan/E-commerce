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

func TestHandler_CreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		callerID   int64
		setup      func(orderSvc *MockOrderService)
		errCode    int
		errMessage string
	}{
		{
			name:     "invalid body",
			body:     `{"items":`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "empty items",
			body:     `{"items":[]}`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "product not found",
			body:     `{"items":[{"product_id":99,"quantity":1}]}`,
			callerID: 1,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.CreateFn = func(_ context.Context, _ int64, _ *models.CreateOrder) (*models.Order, error) {
					return nil, apperrors.NotFound("product not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "product not found",
		},
		{
			name:     "success",
			body:     `{"items":[{"product_id":1,"quantity":2}]}`,
			callerID: 1,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.CreateFn = func(_ context.Context, userID int64, _ *models.CreateOrder) (*models.Order, error) {
					return &models.Order{ID: 10, UserID: userID, Status: models.OrderStatusPending}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := &MockOrderService{}
			if tt.setup != nil {
				tt.setup(orderSvc)
			}
			h := newTestHandler(&service.Services{Order: orderSvc})

			c, w := newTestContext(http.MethodPost, "/orders", tt.body)
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.CreateOrder(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

func TestHandler_GetOrderByID(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		callerID   int64
		callerRole models.Role
		setup      func(orderSvc *MockOrderService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			callerID:   1,
			callerRole: models.RoleCustomer,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid order id",
		},
		{
			name:       "order not found",
			idParam:    "99",
			callerID:   1,
			callerRole: models.RoleCustomer,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, _ int64) (*models.Order, error) {
					return nil, apperrors.NotFound("order not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "order not found",
		},
		{
			name:       "forbidden — not owner and not admin",
			idParam:    "1",
			callerID:   99,
			callerRole: models.RoleCustomer,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1}, nil // owner is userID=1
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "access denied",
		},
		{
			name:       "success as owner",
			idParam:    "1",
			callerID:   1,
			callerRole: models.RoleCustomer,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1}, nil
				}
			},
		},
		{
			name:       "success as admin viewing other user's order",
			idParam:    "1",
			callerID:   99,
			callerRole: models.RoleAdmin,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := &MockOrderService{}
			if tt.setup != nil {
				tt.setup(orderSvc)
			}
			h := newTestHandler(&service.Services{Order: orderSvc})

			c, w := newTestContext(http.MethodGet, "/orders/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, tt.callerRole)
			h.GetOrderByID(c)

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

func TestHandler_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		body       string
		setup      func(orderSvc *MockOrderService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			body:       `{"status":"confirmed"}`,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid order id",
		},
		{
			name:    "invalid status in body",
			idParam: "1",
			body:    `{"status":"unknown"}`,
			errCode: http.StatusBadRequest,
		},
		{
			name:    "order not found",
			idParam: "99",
			body:    `{"status":"confirmed"}`,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.UpdateStatusFn = func(_ context.Context, _ int64, _ models.OrderStatus) error {
					return apperrors.NotFound("order not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "order not found",
		},
		{
			name:    "success",
			idParam: "1",
			body:    `{"status":"confirmed"}`,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.UpdateStatusFn = func(_ context.Context, id int64, s models.OrderStatus) error {
					assert.Equal(t, int64(1), id)
					assert.Equal(t, models.OrderStatusConfirmed, s)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := &MockOrderService{}
			if tt.setup != nil {
				tt.setup(orderSvc)
			}
			h := newTestHandler(&service.Services{Order: orderSvc})

			c, w := newTestContext(http.MethodPatch, "/orders/"+tt.idParam+"/status", tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.UpdateOrderStatus(c)

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

func TestHandler_CancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		callerID   int64
		setup      func(orderSvc *MockOrderService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid order id",
		},
		{
			name:     "order not found",
			idParam:  "99",
			callerID: 1,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, _ int64) (*models.Order, error) {
					return nil, apperrors.NotFound("order not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "order not found",
		},
		{
			name:     "forbidden — not owner",
			idParam:  "1",
			callerID: 99,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1}, nil
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "access denied",
		},
		{
			name:     "already cancelled",
			idParam:  "1",
			callerID: 1,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1, Status: models.OrderStatusPending}, nil
				}
				orderSvc.CancelFn = func(_ context.Context, _ int64) error {
					return apperrors.BadRequest("order cannot be cancelled", nil)
				}
			},
			errCode:    http.StatusBadRequest,
			errMessage: "order cannot be cancelled",
		},
		{
			name:     "success",
			idParam:  "1",
			callerID: 1,
			setup: func(orderSvc *MockOrderService) {
				orderSvc.GetByIDFn = func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, UserID: 1, Status: models.OrderStatusPending}, nil
				}
				orderSvc.CancelFn = func(_ context.Context, _ int64) error { return nil }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := &MockOrderService{}
			if tt.setup != nil {
				tt.setup(orderSvc)
			}
			h := newTestHandler(&service.Services{Order: orderSvc})

			c, w := newTestContext(http.MethodPatch, "/orders/"+tt.idParam+"/cancel", "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.CancelOrder(c)

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
