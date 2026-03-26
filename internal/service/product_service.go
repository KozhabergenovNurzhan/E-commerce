package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/utils"
)

type ProductService interface {
	Create(ctx context.Context, sellerID *int64, req *domain.CreateProductRequest) (*domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	Update(ctx context.Context, id int64, req *domain.UpdateProductRequest) (*domain.Product, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error)
	ListBySeller(ctx context.Context, sellerID int64, f *domain.ProductFilter) ([]*domain.Product, int, error)
	ListCategories(ctx context.Context) ([]*domain.Category, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, sellerID *int64, req *domain.CreateProductRequest) (*domain.Product, error) {
	now := utils.Now()
	p := &domain.Product{
		CategoryID:  req.CategoryID,
		SellerID:    sellerID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productService) Update(ctx context.Context, id int64, req *domain.UpdateProductRequest) (*domain.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Name = req.Name
	p.Description = req.Description
	p.Price = req.Price
	p.Stock = req.Stock
	p.ImageURL = req.ImageURL
	p.UpdatedAt = utils.Now()
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) List(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return s.repo.List(ctx, f)
}

func (s *productService) ListBySeller(ctx context.Context, sellerID int64, f *domain.ProductFilter) ([]*domain.Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return s.repo.ListBySeller(ctx, sellerID, f)
}

func (s *productService) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.repo.ListCategories(ctx)
}
