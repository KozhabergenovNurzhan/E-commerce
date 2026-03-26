package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type OrderService interface {
	Create(ctx context.Context, userID uuid.UUID, req *domain.CreateOrderRequest) (*domain.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.Order, int, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) OrderService {
	return &orderService{orderRepo: orderRepo, productRepo: productRepo}
}

func (s *orderService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateOrderRequest) (*domain.Order, error) {
	var total float64
	items := make([]domain.OrderItem, 0, len(req.Items))

	for _, r := range req.Items {
		product, err := s.productRepo.FindByID(ctx, r.ProductID)
		if err != nil {
			return nil, err
		}
		if product.Stock < r.Quantity {
			return nil, apperrors.ErrBadRequest
		}
		total += product.Price * float64(r.Quantity)
		items = append(items, domain.OrderItem{
			ID:        uuid.New(),
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
			UnitPrice: product.Price,
		})
	}

	now := time.Now().UTC()
	order := &domain.Order{
		ID:         uuid.New(),
		UserID:     userID,
		Status:     domain.OrderStatusPending,
		TotalPrice: total,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	for i := range items {
		items[i].OrderID = order.ID
	}
	order.Items = items

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *orderService) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.orderRepo.ListByUser(ctx, userID, limit, (page-1)*limit)
}

func (s *orderService) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != domain.OrderStatusPending {
		return apperrors.ErrBadRequest
	}
	return s.orderRepo.UpdateStatus(ctx, id, domain.OrderStatusCancelled)
}
