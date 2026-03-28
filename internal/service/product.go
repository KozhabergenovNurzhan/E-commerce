package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, sellerID *int64, req *models.CreateProduct) (*models.Product, error) {
	now := utils.Now()
	p := &models.Product{
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

func (s *ProductService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id int64, req *models.UpdateProduct) (*models.Product, error) {
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

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
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

func (s *ProductService) ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error) {
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

func (s *ProductService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *ProductService) CreateCategory(ctx context.Context, req *models.CreateCategory) (*models.Category, error) {
	c := &models.Category{
		Name:      req.Name,
		Slug:      req.Slug,
		CreatedAt: utils.Now(),
	}

	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, id int64, req *models.UpdateCategory) (*models.Category, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	var target *models.Category
	for _, cat := range cats {
		if cat.ID == id {
			target = cat
			break
		}
	}
	if target == nil {
		return nil, apperrors.NotFound("category not found", nil)
	}

	target.Name = req.Name
	target.Slug = req.Slug

	if err := s.repo.UpdateCategory(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, id int64) error {
	return s.repo.DeleteCategory(ctx, id)
}
