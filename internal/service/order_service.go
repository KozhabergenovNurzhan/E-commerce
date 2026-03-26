package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/utils"
)

type OrderService interface {
	Create(ctx context.Context, userID int64, req *domain.CreateOrderRequest) (*domain.Order, error)
	GetByID(ctx context.Context, id int64) (*domain.Order, error)
	ListByUser(ctx context.Context, userID int64, page, limit int) ([]*domain.Order, int, error)
	Cancel(ctx context.Context, id int64) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) OrderService {
	return &orderService{orderRepo: orderRepo, productRepo: productRepo}
}

func (s *orderService) Create(ctx context.Context, userID int64, req *domain.CreateOrderRequest) (*domain.Order, error) {
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
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
			UnitPrice: product.Price,
		})
	}

	now := utils.Now()
	order := &domain.Order{
		UserID:     userID,
		Status:     domain.OrderStatusPending,
		TotalPrice: total,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      items,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *orderService) ListByUser(ctx context.Context, userID int64, page, limit int) ([]*domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.orderRepo.ListByUser(ctx, userID, limit, (page-1)*limit)
}

func (s *orderService) Cancel(ctx context.Context, id int64) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != domain.OrderStatusPending {
		return apperrors.ErrBadRequest
	}
	return s.orderRepo.UpdateStatus(ctx, id, domain.OrderStatusCancelled)
}
