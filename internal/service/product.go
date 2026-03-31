package service

import (
	"context"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/cache"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type ProductService struct {
	repo  repository.ProductRepository
	cache *cache.RedisCache
}

func NewProductService(repo repository.ProductRepository, c *cache.RedisCache) *ProductService {
	return &ProductService{repo: repo, cache: c}
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
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	if s.cache != nil {
		if p, err := s.cache.GetProduct(ctx, id); err == nil {
			return p, nil
		}
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetProduct(ctx, p, 5*time.Minute)
	}

	return p, nil
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

	if s.cache != nil {
		_ = s.cache.DeleteProduct(ctx, id)
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.DeleteProduct(ctx, id)
	}

	return nil
}

func (s *ProductService) List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
	normalizeFilter(f)
	return s.repo.List(ctx, f)
}

func (s *ProductService) ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error) {
	normalizeFilter(f)
	return s.repo.ListBySeller(ctx, sellerID, f)
}

func normalizeFilter(f *models.ProductFilter) {
	if f.Page < 1 {
		f.Page = 1
	}

	if f.Limit < 1 {
		f.Limit = 20
	} else if f.Limit > 100 {
		f.Limit = 100
	}
}

func (s *ProductService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	if s.cache != nil {
		if cats, err := s.cache.GetCategories(ctx); err == nil {
			return cats, nil
		}
	}

	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetCategories(ctx, cats, 10*time.Minute)
	}

	return cats, nil
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

	if s.cache != nil {
		_ = s.cache.DeleteCategories(ctx)
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

	if s.cache != nil {
		_ = s.cache.DeleteCategories(ctx)
	}

	return target, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, id int64) error {
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.DeleteCategories(ctx)
	}

	return nil
}
