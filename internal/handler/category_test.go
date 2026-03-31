package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func TestHandler_CreateCategory(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:    "invalid body",
			body:    `{"name":`,
			errCode: http.StatusBadRequest,
		},
		{
			name: "slug conflict",
			body: `{"name":"Electronics","slug":"electronics"}`,
			setup: func(svc *MockProductService) {
				svc.CreateCategoryFn = func(_ context.Context, _ *models.CreateCategory) (*models.Category, error) {
					return nil, apperrors.Conflict("slug already exists", nil)
				}
			},
			errCode:    http.StatusConflict,
			errMessage: "slug already exists",
		},
		{
			name: "success",
			body: `{"name":"Electronics","slug":"electronics"}`,
			setup: func(svc *MockProductService) {
				svc.CreateCategoryFn = func(_ context.Context, req *models.CreateCategory) (*models.Category, error) {
					return &models.Category{ID: 1, Name: req.Name, Slug: req.Slug}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productSvc := &MockProductService{}
			if tt.setup != nil {
				tt.setup(productSvc)
			}
			h := newTestHandler(&service.Services{Product: productSvc})

			c, w := newTestContext(http.MethodPost, "/categories", tt.body)
			h.CreateCategory(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusCreated, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_UpdateCategory(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		body       string
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			body:       `{"name":"Electronics","slug":"electronics"}`,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid category id",
		},
		{
			name:    "invalid body",
			idParam: "1",
			body:    `{"name":`,
			errCode: http.StatusBadRequest,
		},
		{
			name:    "not found",
			idParam: "99",
			body:    `{"name":"Electronics","slug":"electronics"}`,
			setup: func(svc *MockProductService) {
				svc.UpdateCategoryFn = func(_ context.Context, _ int64, _ *models.UpdateCategory) (*models.Category, error) {
					return nil, apperrors.NotFound("category not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "category not found",
		},
		{
			name:    "success",
			idParam: "1",
			body:    `{"name":"Electronics v2","slug":"electronics-v2"}`,
			setup: func(svc *MockProductService) {
				svc.UpdateCategoryFn = func(_ context.Context, id int64, req *models.UpdateCategory) (*models.Category, error) {
					return &models.Category{ID: id, Name: req.Name, Slug: req.Slug}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productSvc := &MockProductService{}
			if tt.setup != nil {
				tt.setup(productSvc)
			}
			h := newTestHandler(&service.Services{Product: productSvc})

			c, w := newTestContext(http.MethodPut, "/categories/"+tt.idParam, tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.UpdateCategory(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestHandler_DeleteCategory(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			errCode:    http.StatusBadRequest,
			errMessage: "invalid category id",
		},
		{
			name:    "not found",
			idParam: "99",
			setup: func(svc *MockProductService) {
				svc.DeleteCategoryFn = func(_ context.Context, _ int64) error {
					return apperrors.NotFound("category not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "category not found",
		},
		{
			name:    "success",
			idParam: "1",
			setup: func(svc *MockProductService) {
				svc.DeleteCategoryFn = func(_ context.Context, id int64) error {
					assert.Equal(t, int64(1), id)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productSvc := &MockProductService{}
			if tt.setup != nil {
				tt.setup(productSvc)
			}
			h := newTestHandler(&service.Services{Product: productSvc})

			c, w := newTestContext(http.MethodDelete, "/categories/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.DeleteCategory(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusNoContent, w.Code)
			assert.Empty(t, w.Body.String())
		})
	}
}
