package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type CartService struct {
	repo        repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(repo repository.CartRepository, productRepo repository.ProductRepository) *CartService {
	return &CartService{repo: repo, productRepo: productRepo}
}

func (s *CartService) AddItem(ctx context.Context, userID int64, req *models.AddToCart) error {
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	if product.Stock < req.Quantity {
		return apperrors.BadRequest("insufficient stock", nil)
	}

	now := utils.Now()
	item := &models.CartItemRecord{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repo.Upsert(ctx, item)
}

func (s *CartService) GetCart(ctx context.Context, userID int64) (*models.Cart, error) {
	items, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var total float64
	respItems := make([]*models.CartItem, 0, len(items))
	for _, item := range items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			continue // product removed from catalog
		}
		subtotal := product.Price * float64(item.Quantity)
		total += subtotal
		respItems = append(respItems, &models.CartItem{
			ID:       item.ID,
			Product:  product,
			Quantity: item.Quantity,
			Subtotal: subtotal,
		})
	}

	return &models.Cart{Items: respItems, TotalPrice: total}, nil
}

func (s *CartService) UpdateItem(ctx context.Context, userID, productID int64, req *models.UpdateCartItem) error {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < req.Quantity {
		return apperrors.BadRequest("insufficient stock", nil)
	}

	now := utils.Now()
	item := &models.CartItemRecord{
		UserID:    userID,
		ProductID: productID,
		Quantity:  req.Quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repo.Upsert(ctx, item)
}

func (s *CartService) RemoveItem(ctx context.Context, userID, productID int64) error {
	return s.repo.Delete(ctx, userID, productID)
}

func (s *CartService) Clear(ctx context.Context, userID int64) error {
	return s.repo.Clear(ctx, userID)
}
