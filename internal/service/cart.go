package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/utils"
)

type CartService interface {
	AddItem(ctx context.Context, userID int64, req *domain.AddToCartRequest) error
	GetCart(ctx context.Context, userID int64) (*domain.CartResponse, error)
	UpdateItem(ctx context.Context, userID, productID int64, req *domain.UpdateCartItemRequest) error
	RemoveItem(ctx context.Context, userID, productID int64) error
	Clear(ctx context.Context, userID int64) error
}

type cartService struct {
	repo        repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(repo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{repo: repo, productRepo: productRepo}
}

func (s *cartService) AddItem(ctx context.Context, userID int64, req *domain.AddToCartRequest) error {
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	if product.Stock < req.Quantity {
		return apperrors.ErrBadRequest
	}

	now := utils.Now()
	item := &domain.CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.Upsert(ctx, item)
}

func (s *cartService) GetCart(ctx context.Context, userID int64) (*domain.CartResponse, error) {
	items, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var total float64
	respItems := make([]*domain.CartItemResponse, 0, len(items))
	for _, item := range items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			continue // product removed from catalog
		}
		subtotal := product.Price * float64(item.Quantity)
		total += subtotal
		respItems = append(respItems, &domain.CartItemResponse{
			ID:       item.ID,
			Product:  product,
			Quantity: item.Quantity,
			Subtotal: subtotal,
		})
	}

	return &domain.CartResponse{Items: respItems, TotalPrice: total}, nil
}

func (s *cartService) UpdateItem(ctx context.Context, userID, productID int64, req *domain.UpdateCartItemRequest) error {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < req.Quantity {
		return apperrors.ErrBadRequest
	}

	now := utils.Now()
	item := &domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  req.Quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.Upsert(ctx, item)
}

func (s *cartService) RemoveItem(ctx context.Context, userID, productID int64) error {
	return s.repo.Delete(ctx, userID, productID)
}

func (s *cartService) Clear(ctx context.Context, userID int64) error {
	return s.repo.Clear(ctx, userID)
}
