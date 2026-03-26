package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		req         *domain.CreateOrderRequest
		findProduct func(ctx context.Context, id int64) (*domain.Product, error)
		createOrder func(ctx context.Context, order *domain.Order) error
		wantErr     error
		check       func(t *testing.T, order *domain.Order)
	}{
		{
			name: "success — multiple items",
			req: &domain.CreateOrderRequest{
				Items: []domain.CreateOrderItemRequest{
					{ProductID: 1, Quantity: 2},
					{ProductID: 2, Quantity: 1},
				},
			},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Price: 50.0, Stock: 10}, nil
			},
			createOrder: func(_ context.Context, order *domain.Order) error {
				order.ID = 1
				return nil
			},
			check: func(t *testing.T, order *domain.Order) {
				assert.Equal(t, int64(1), order.ID)
				assert.Equal(t, 150.0, order.TotalPrice) // 2*50 + 1*50
				assert.Equal(t, domain.OrderStatusPending, order.Status)
				assert.Len(t, order.Items, 2)
			},
		},
		{
			name: "insufficient stock",
			req: &domain.CreateOrderRequest{
				Items: []domain.CreateOrderItemRequest{
					{ProductID: 1, Quantity: 5},
				},
			},
			findProduct: func(_ context.Context, id int64) (*domain.Product, error) {
				return &domain.Product{ID: id, Price: 50.0, Stock: 1}, nil
			},
			wantErr: apperrors.ErrBadRequest,
		},
		{
			name: "product not found",
			req: &domain.CreateOrderRequest{
				Items: []domain.CreateOrderItemRequest{
					{ProductID: 99, Quantity: 1},
				},
			},
			findProduct: func(_ context.Context, _ int64) (*domain.Product, error) {
				return nil, apperrors.ErrNotFound
			},
			wantErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.findProduct}
			orderRepo := &testutil.MockOrderRepo{CreateFn: tt.createOrder}

			order, err := service.NewOrderService(orderRepo, productRepo).
				Create(context.Background(), 1, tt.req)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, order)
		})
	}
}

func TestCancelOrder(t *testing.T) {
	tests := []struct {
		name    string
		status  domain.OrderStatus
		wantErr error
	}{
		{name: "pending order can be cancelled", status: domain.OrderStatusPending},
		{name: "confirmed order cannot be cancelled", status: domain.OrderStatusConfirmed, wantErr: apperrors.ErrBadRequest},
		{name: "shipping order cannot be cancelled", status: domain.OrderStatusShipping, wantErr: apperrors.ErrBadRequest},
		{name: "delivered order cannot be cancelled", status: domain.OrderStatusDelivered, wantErr: apperrors.ErrBadRequest},
		{name: "already cancelled order cannot be cancelled", status: domain.OrderStatusCancelled, wantErr: apperrors.ErrBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := &testutil.MockOrderRepo{
				FindByIDFn: func(_ context.Context, id int64) (*domain.Order, error) {
					return &domain.Order{ID: id, Status: tt.status}, nil
				},
				UpdateStatusFn: func(_ context.Context, _ int64, _ domain.OrderStatus) error {
					return nil
				},
			}

			err := service.NewOrderService(orderRepo, nil).Cancel(context.Background(), 1)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name      string
		findOrder func(ctx context.Context, id int64) (*domain.Order, error)
		newStatus domain.OrderStatus
		wantErr   error
	}{
		{
			name: "success",
			findOrder: func(_ context.Context, id int64) (*domain.Order, error) {
				return &domain.Order{ID: id, Status: domain.OrderStatusPending}, nil
			},
			newStatus: domain.OrderStatusConfirmed,
		},
		{
			name: "order not found",
			findOrder: func(_ context.Context, _ int64) (*domain.Order, error) {
				return nil, apperrors.ErrNotFound
			},
			newStatus: domain.OrderStatusConfirmed,
			wantErr:   apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := &testutil.MockOrderRepo{
				FindByIDFn:     tt.findOrder,
				UpdateStatusFn: func(_ context.Context, _ int64, _ domain.OrderStatus) error { return nil },
			}

			err := service.NewOrderService(orderRepo, nil).UpdateStatus(context.Background(), 1, tt.newStatus)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
