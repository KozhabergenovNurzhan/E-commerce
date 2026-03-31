package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestProductService_Create(t *testing.T) {
	ctx := context.Background()
	sellerID := int64(42)

	tests := []struct {
		name       string
		sellerID   *int64
		req        *models.CreateProduct
		setup      func(r *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, p *models.Product)
	}{
		{
			name:     "success",
			sellerID: &sellerID,
			req:      &models.CreateProduct{CategoryID: 1, Name: "Widget", Price: 9.99, Stock: 50},
			setup: func(r *MockProductRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.Product")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.Product).ID = 7
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, p *models.Product) {
				assert.Equal(t, int64(7), p.ID)
				assert.Equal(t, "Widget", p.Name)
				assert.Equal(t, 9.99, p.Price)
				assert.Equal(t, &sellerID, p.SellerID)
				assert.False(t, p.CreatedAt.IsZero())
			},
		},
		{
			name:     "repo error",
			sellerID: nil,
			req:      &models.CreateProduct{Name: "x", Price: 1, CategoryID: 1},
			setup: func(r *MockProductRepo) {
				r.On("Create", ctx, mock.AnythingOfType("*models.Product")).
					Return(apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(productRepo)
			}

			svc := service.NewProductService(productRepo, nil)
			p, err := svc.Create(ctx, tt.sellerID, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, p)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         int64
		setup      func(r *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, p *models.Product)
	}{
		{
			name: "success",
			id:   1,
			setup: func(r *MockProductRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Name: "Widget"}, nil).Once()
			},
			check: func(t *testing.T, p *models.Product) {
				assert.Equal(t, int64(1), p.ID)
				assert.Equal(t, "Widget", p.Name)
			},
		},
		{
			name: "not found",
			id:   99,
			setup: func(r *MockProductRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(productRepo)
			}

			svc := service.NewProductService(productRepo, nil)
			p, err := svc.GetByID(ctx, tt.id)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, p)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         int64
		req        *models.UpdateProduct
		setup      func(r *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, p *models.Product)
	}{
		{
			name: "success — fields are applied",
			id:   1,
			req:  &models.UpdateProduct{Name: "New", Price: 1.0},
			setup: func(r *MockProductRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1, Name: "Old"}, nil).Once()
				r.On("Update", ctx, mock.AnythingOfType("*models.Product")).
					Return(nil).Once()
			},
			check: func(t *testing.T, p *models.Product) {
				assert.Equal(t, "New", p.Name)
				assert.False(t, p.UpdatedAt.IsZero())
			},
		},
		{
			name: "product not found",
			id:   99,
			req:  &models.UpdateProduct{Name: "x", Price: 1},
			setup: func(r *MockProductRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
		{
			name: "repo update error",
			id:   1,
			req:  &models.UpdateProduct{Name: "x", Price: 1},
			setup: func(r *MockProductRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1}, nil).Once()
				r.On("Update", ctx, mock.AnythingOfType("*models.Product")).
					Return(apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(productRepo)
			}

			svc := service.NewProductService(productRepo, nil)
			p, err := svc.Update(ctx, tt.id, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, p)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		id         int64
		setup      func(r *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name: "success",
			id:   1,
			setup: func(r *MockProductRepo) {
				r.On("Delete", ctx, int64(1)).Return(nil).Once()
			},
		},
		{
			name: "not found",
			id:   99,
			setup: func(r *MockProductRepo) {
				r.On("Delete", ctx, int64(99)).
					Return(apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(productRepo)
			}

			svc := service.NewProductService(productRepo, nil)
			err := svc.Delete(ctx, tt.id)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_List_PaginationDefaults(t *testing.T) {
	ctx := context.Background()

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
			productRepo := new(MockProductRepo)
			productRepo.On("List", ctx, mock.MatchedBy(func(f *models.ProductFilter) bool {
				return f.Page == tt.wantPage && f.Limit == tt.wantLimit
			})).Return([]*models.Product{}, 0, nil).Once()

			f := tt.input
			_, _, err := service.NewProductService(productRepo, nil).List(ctx, &f)

			require.NoError(t, err)
			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_ListCategories(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(r *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, cats []*models.Category)
	}{
		{
			name: "success",
			setup: func(r *MockProductRepo) {
				r.On("ListCategories", ctx).Return([]*models.Category{
					{ID: 1, Name: "Electronics"},
					{ID: 2, Name: "Books"},
				}, nil).Once()
			},
			check: func(t *testing.T, cats []*models.Category) {
				assert.Len(t, cats, 2)
				assert.Equal(t, "Electronics", cats[0].Name)
			},
		},
		{
			name: "repo error",
			setup: func(r *MockProductRepo) {
				r.On("ListCategories", ctx).
					Return(nil, apperrors.Internal("internal server error", errDB)).Once()
			},
			errCode:    http.StatusInternalServerError,
			errMsg:     "internal server error",
			errWrapped: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(productRepo)
			}

			svc := service.NewProductService(productRepo, nil)
			cats, err := svc.ListCategories(ctx)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, cats)
			}

			productRepo.AssertExpectations(t)
		})
	}
}
