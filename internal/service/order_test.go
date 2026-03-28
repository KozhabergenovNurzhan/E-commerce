package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
)

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		req         *models.CreateOrder
		findProduct func(ctx context.Context, id int64) (*models.Product, error)
		createOrder func(ctx context.Context, order *models.Order) error
		wantCode    int
		check       func(t *testing.T, order *models.Order)
	}{
		{
			name: "success — multiple items",
			req: &models.CreateOrder{
				Items: []models.CreateOrderItem{
					{ProductID: 1, Quantity: 2},
					{ProductID: 2, Quantity: 1},
				},
			},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Price: 50.0, Stock: 10}, nil
			},
			createOrder: func(_ context.Context, order *models.Order) error {
				order.ID = 1
				return nil
			},
			check: func(t *testing.T, order *models.Order) {
				assert.Equal(t, int64(1), order.ID)
				assert.Equal(t, 150.0, order.TotalPrice) // 2*50 + 1*50
				assert.Equal(t, models.OrderStatusPending, order.Status)
				assert.Len(t, order.Items, 2)
			},
		},
		{
			name: "insufficient stock",
			req: &models.CreateOrder{
				Items: []models.CreateOrderItem{
					{ProductID: 1, Quantity: 5},
				},
			},
			findProduct: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Price: 50.0, Stock: 1}, nil
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "product not found",
			req: &models.CreateOrder{
				Items: []models.CreateOrderItem{
					{ProductID: 99, Quantity: 1},
				},
			},
			findProduct: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}
			orderRepo := &testutil.MockOrderRepo{CreateFn: tt.createOrder}

			order, err := service.NewOrderService(nil, orderRepo, productRepo).
				Create(context.Background(), 1, tt.req)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
			tt.check(t, order)
		})
	}
}

func TestCancelOrder(t *testing.T) {
	tests := []struct {
		name     string
		status   models.OrderStatus
		wantCode int
	}{
		{name: "pending order can be cancelled", status: models.OrderStatusPending},
		{name: "confirmed order cannot be cancelled", status: models.OrderStatusConfirmed, wantCode: http.StatusBadRequest},
		{name: "shipping order cannot be cancelled", status: models.OrderStatusShipping, wantCode: http.StatusBadRequest},
		{name: "delivered order cannot be cancelled", status: models.OrderStatusDelivered, wantCode: http.StatusBadRequest},
		{name: "already cancelled order cannot be cancelled", status: models.OrderStatusCancelled, wantCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := &testutil.MockOrderRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Order, error) {
					return &models.Order{ID: id, Status: tt.status}, nil
				},
				UpdateStatusFn: func(_ context.Context, _ int64, _ models.OrderStatus) error {
					return nil
				},
			}

			err := service.NewOrderService(nil, orderRepo, nil).Cancel(context.Background(), 1)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name      string
		findOrder func(ctx context.Context, id int64) (*models.Order, error)
		newStatus models.OrderStatus
		wantCode  int
	}{
		{
			name: "success",
			findOrder: func(_ context.Context, id int64) (*models.Order, error) {
				return &models.Order{ID: id, Status: models.OrderStatusPending}, nil
			},
			newStatus: models.OrderStatusConfirmed,
		},
		{
			name: "order not found",
			findOrder: func(_ context.Context, _ int64) (*models.Order, error) {
				return nil, apperrors.NotFound("order not found", nil)
			},
			newStatus: models.OrderStatusConfirmed,
			wantCode:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := &testutil.MockOrderRepo{
				FindByIDFn:     tt.findOrder,
				UpdateStatusFn: func(_ context.Context, _ int64, _ models.OrderStatus) error { return nil },
			}

			err := service.NewOrderService(nil, orderRepo, nil).UpdateStatus(context.Background(), 1, tt.newStatus)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
