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

func TestOrderService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		req        *models.CreateOrder
		setup      func(orderRepo *MockOrderRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, o *models.Order)
	}{
		{
			name: "success — multiple items",
			req: &models.CreateOrder{Items: []models.CreateOrderItem{
				{ProductID: 1, Quantity: 2},
				{ProductID: 2, Quantity: 1},
			}},
			setup: func(orderRepo *MockOrderRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Price: 50.0, Stock: 10}, nil).Once()
				productRepo.On("FindByID", ctx, int64(2)).
					Return(&models.Product{ID: 2, Price: 50.0, Stock: 10}, nil).Once()
				orderRepo.On("Create", ctx, mock.AnythingOfType("*models.Order")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.Order).ID = 1
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, o *models.Order) {
				assert.Equal(t, int64(1), o.ID)
				assert.Equal(t, 150.0, o.TotalPrice) // 2*50 + 1*50
				assert.Equal(t, models.OrderStatusPending, o.Status)
				assert.Len(t, o.Items, 2)
			},
		},
		{
			name: "insufficient stock",
			req: &models.CreateOrder{Items: []models.CreateOrderItem{
				{ProductID: 1, Quantity: 5},
			}},
			setup: func(_ *MockOrderRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Price: 50.0, Stock: 1}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "insufficient stock",
		},
		{
			name: "product not found",
			req: &models.CreateOrder{Items: []models.CreateOrderItem{
				{ProductID: 99, Quantity: 1},
			}},
			setup: func(_ *MockOrderRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := new(MockOrderRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(orderRepo, productRepo)
			}

			svc := service.NewOrderService(nil, orderRepo, productRepo)
			order, err := svc.Create(ctx, 1, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, order)
			}

			orderRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_Cancel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		status     models.OrderStatus
		setup      func(orderRepo *MockOrderRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:   "pending order can be cancelled",
			status: models.OrderStatusPending,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Order{ID: 1, Status: models.OrderStatusPending}, nil).Once()
				r.On("UpdateStatus", ctx, int64(1), models.OrderStatusCancelled).
					Return(nil).Once()
			},
		},
		{
			name:   "confirmed order cannot be cancelled",
			status: models.OrderStatusConfirmed,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Order{ID: 1, Status: models.OrderStatusConfirmed}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "only pending orders can be cancelled",
		},
		{
			name:   "already cancelled order cannot be cancelled",
			status: models.OrderStatusCancelled,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Order{ID: 1, Status: models.OrderStatusCancelled}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "only pending orders can be cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := new(MockOrderRepo)
			if tt.setup != nil {
				tt.setup(orderRepo)
			}

			svc := service.NewOrderService(nil, orderRepo, nil)
			err := svc.Cancel(ctx, 1)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			orderRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         int64
		newStatus  models.OrderStatus
		setup      func(r *MockOrderRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:      "success — pending to confirmed",
			id:        1,
			newStatus: models.OrderStatusConfirmed,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Order{ID: 1, Status: models.OrderStatusPending}, nil).Once()
				r.On("UpdateStatus", ctx, int64(1), models.OrderStatusConfirmed).
					Return(nil).Once()
			},
		},
		{
			name:      "invalid transition — pending to shipping",
			id:        1,
			newStatus: models.OrderStatusShipping,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Order{ID: 1, Status: models.OrderStatusPending}, nil).Once()
			},
			errCode: http.StatusBadRequest,
			errMsg:  "cannot transition from pending to shipping",
		},
		{
			name:      "order not found",
			id:        99,
			newStatus: models.OrderStatusConfirmed,
			setup: func(r *MockOrderRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("order not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := new(MockOrderRepo)
			if tt.setup != nil {
				tt.setup(orderRepo)
			}

			svc := service.NewOrderService(nil, orderRepo, nil)
			err := svc.UpdateStatus(ctx, tt.id, tt.newStatus)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			orderRepo.AssertExpectations(t)
		})
	}
}
