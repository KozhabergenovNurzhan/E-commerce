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

func TestProductCreate(t *testing.T) {
	sellerID := int64(42)
	req := &models.CreateProduct{
		CategoryID:  1,
		Name:        "Widget",
		Description: "A fine widget",
		Price:       9.99,
		Stock:       50,
		ImageURL:    "http://example.com/widget.png",
	}

	productRepo := &testutil.MockProductRepo{
		CreateFn: func(_ context.Context, p *models.Product) error {
			p.ID = 7
			return nil
		},
	}

	p, err := service.NewProductService(productRepo, nil).Create(context.Background(), &sellerID, req)

	require.NoError(t, err)
	assert.Equal(t, int64(7), p.ID)
	assert.Equal(t, req.Name, p.Name)
	assert.Equal(t, req.Price, p.Price)
	assert.Equal(t, req.Stock, p.Stock)
	assert.True(t, p.IsActive)
	assert.Equal(t, &sellerID, p.SellerID)
	assert.False(t, p.CreatedAt.IsZero())
	assert.False(t, p.UpdatedAt.IsZero())
}

func TestProductCreate_RepoError(t *testing.T) {
	productRepo := &testutil.MockProductRepo{
		CreateFn: func(_ context.Context, _ *models.Product) error {
			return apperrors.Internal("internal server error", nil)
		},
	}

	_, err := service.NewProductService(productRepo, nil).
		Create(context.Background(), nil, &models.CreateProduct{Name: "x", Price: 1, CategoryID: 1})

	assertCode(t, err, http.StatusInternalServerError)
}

func TestProductGetByID(t *testing.T) {
	tests := []struct {
		name     string
		stub     func(ctx context.Context, id int64) (*models.Product, error)
		wantCode int
	}{
		{
			name: "found",
			stub: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Name: "Widget"}, nil
			},
		},
		{
			name: "not found",
			stub: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &testutil.MockProductRepo{FindByIDFn: tt.stub}
			p, err := service.NewProductService(productRepo, nil).GetByID(context.Background(), 1)
			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, int64(1), p.ID)
		})
	}
}

func TestProductUpdate(t *testing.T) {
	tests := []struct {
		name      string
		findByID  func(ctx context.Context, id int64) (*models.Product, error)
		updateFn  func(ctx context.Context, p *models.Product) error
		wantCode  int
		checkName string
	}{
		{
			name: "success — fields are applied",
			findByID: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id, Name: "Old"}, nil
			},
			updateFn:  func(_ context.Context, _ *models.Product) error { return nil },
			checkName: "New",
		},
		{
			name: "product not found",
			findByID: func(_ context.Context, _ int64) (*models.Product, error) {
				return nil, apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "repo update error",
			findByID: func(_ context.Context, id int64) (*models.Product, error) {
				return &models.Product{ID: id}, nil
			},
			updateFn: func(_ context.Context, _ *models.Product) error {
				return apperrors.Internal("internal server error", nil)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &testutil.MockProductRepo{
				FindByIDFn: tt.findByID,
				UpdateFn:   tt.updateFn,
			}
			req := &models.UpdateProduct{Name: "New", Price: 1.0}
			p, err := service.NewProductService(productRepo, nil).Update(context.Background(), 1, req)

			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.checkName, p.Name)
			assert.False(t, p.UpdatedAt.IsZero())
		})
	}
}

func TestProductDelete(t *testing.T) {
	tests := []struct {
		name     string
		stub     func(ctx context.Context, id int64) error
		wantCode int
	}{
		{
			name: "success",
			stub: func(_ context.Context, _ int64) error { return nil },
		},
		{
			name: "not found",
			stub: func(_ context.Context, _ int64) error {
				return apperrors.NotFound("product not found", nil)
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := &testutil.MockProductRepo{DeleteFn: tt.stub}
			err := service.NewProductService(productRepo, nil).Delete(context.Background(), 1)
			if tt.wantCode != 0 {
				assertCode(t, err, tt.wantCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductList_PaginationDefaults(t *testing.T) {
	tests := []struct {
		name      string
		input     models.ProductFilter
		wantPage  int
		wantLimit int
	}{
		{name: "zero page and limit use defaults", input: models.ProductFilter{Page: 0, Limit: 0}, wantPage: 1, wantLimit: 20},
		{name: "negative page uses default", input: models.ProductFilter{Page: -1, Limit: 10}, wantPage: 1, wantLimit: 10},
		{name: "limit over 100 is capped", input: models.ProductFilter{Page: 1, Limit: 200}, wantPage: 1, wantLimit: 100},
		{name: "valid values are preserved", input: models.ProductFilter{Page: 3, Limit: 50}, wantPage: 3, wantLimit: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFilter *models.ProductFilter
			productRepo := &testutil.MockProductRepo{
				ListFn: func(_ context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
					capturedFilter = f
					return []*models.Product{}, 0, nil
				},
			}
			f := tt.input
			_, _, err := service.NewProductService(productRepo, nil).List(context.Background(), &f)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, capturedFilter.Page)
			assert.Equal(t, tt.wantLimit, capturedFilter.Limit)
		})
	}
}

func TestProductListBySeller_PaginationDefaults(t *testing.T) {
	tests := []struct {
		name      string
		input     models.ProductFilter
		wantPage  int
		wantLimit int
	}{
		{name: "zero values use defaults", input: models.ProductFilter{}, wantPage: 1, wantLimit: 20},
		{name: "limit over 100 is capped", input: models.ProductFilter{Page: 1, Limit: 150}, wantPage: 1, wantLimit: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFilter *models.ProductFilter
			productRepo := &testutil.MockProductRepo{
				ListBySellerFn: func(_ context.Context, _ int64, f *models.ProductFilter) ([]*models.Product, int, error) {
					capturedFilter = f
					return []*models.Product{}, 0, nil
				},
			}
			f := tt.input
			_, _, err := service.NewProductService(productRepo, nil).ListBySeller(context.Background(), 1, &f)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, capturedFilter.Page)
			assert.Equal(t, tt.wantLimit, capturedFilter.Limit)
		})
	}
}

func TestProductListCategories(t *testing.T) {
	cats := []*models.Category{
		{ID: 1, Name: "Electronics"},
		{ID: 2, Name: "Books"},
	}
	productRepo := &testutil.MockProductRepo{
		ListCategoriesFn: func(_ context.Context) ([]*models.Category, error) {
			return cats, nil
		},
	}

	result, err := service.NewProductService(productRepo, nil).ListCategories(context.Background())
	require.NoError(t, err)
	assert.Equal(t, cats, result)
}

func TestProductListCategories_Error(t *testing.T) {
	productRepo := &testutil.MockProductRepo{
		ListCategoriesFn: func(_ context.Context) ([]*models.Category, error) {
			return nil, apperrors.Internal("internal server error", nil)
		},
	}

	_, err := service.NewProductService(productRepo, nil).ListCategories(context.Background())
	assertCode(t, err, http.StatusInternalServerError)
}
