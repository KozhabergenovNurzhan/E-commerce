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

func TestHandler_GetProductByID(t *testing.T) {
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
			errMessage: "invalid product id",
		},
		{
			name:    "not found",
			idParam: "99",
			setup: func(svc *MockProductService) {
				svc.GetByIDFn = func(_ context.Context, _ int64) (*models.Product, error) {
					return nil, apperrors.NotFound("product not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "product not found",
		},
		{
			name:    "success",
			idParam: "1",
			setup: func(svc *MockProductService) {
				svc.GetByIDFn = func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id, Name: "Widget"}, nil
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

			c, w := newTestContext(http.MethodGet, "/products/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.GetProductByID(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		callerID   int64
		callerRole models.Role
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid body",
			body:       `{"name":`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			errCode:    http.StatusBadRequest,
		},
		{
			name:       "service error — invalid category",
			body:       `{"category_id":999,"name":"Widget","price":9.99,"stock":10}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockProductService) {
				svc.CreateFn = func(_ context.Context, _ *int64, _ *models.CreateProduct) (*models.Product, error) {
					return nil, apperrors.BadRequest("invalid category", nil)
				}
			},
			errCode:    http.StatusBadRequest,
			errMessage: "invalid category",
		},
		{
			name:       "success as admin — no seller id",
			body:       `{"category_id":1,"name":"Widget","price":9.99,"stock":10}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockProductService) {
				svc.CreateFn = func(_ context.Context, sellerID *int64, _ *models.CreateProduct) (*models.Product, error) {
					assert.Nil(t, sellerID)
					return &models.Product{ID: 7, Name: "Widget"}, nil
				}
			},
		},
		{
			name:       "success as seller — seller id set automatically",
			body:       `{"category_id":1,"name":"Widget","price":9.99,"stock":10}`,
			callerID:   42,
			callerRole: models.RoleSeller,
			setup: func(svc *MockProductService) {
				svc.CreateFn = func(_ context.Context, sellerID *int64, _ *models.CreateProduct) (*models.Product, error) {
					require.NotNil(t, sellerID)
					assert.Equal(t, int64(42), *sellerID)
					return &models.Product{ID: 8, Name: "Widget"}, nil
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

			c, w := newTestContext(http.MethodPost, "/products", tt.body)
			setAuth(c, tt.callerID, tt.callerRole)
			h.CreateProduct(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

func TestHandler_UpdateProduct(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		body       string
		callerID   int64
		callerRole models.Role
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			body:       `{"name":"Widget","price":9.99,"stock":10}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:       "seller forbidden — not owner",
			idParam:    "1",
			body:       `{"name":"Widget","price":9.99,"stock":10}`,
			callerID:   10,
			callerRole: models.RoleSeller,
			setup: func(svc *MockProductService) {
				ownerID := int64(99)
				svc.GetByIDFn = func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id, SellerID: &ownerID}, nil
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "cannot update another seller's product",
		},
		{
			name:       "success as admin",
			idParam:    "1",
			body:       `{"name":"Updated","price":19.99,"stock":5}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockProductService) {
				svc.UpdateFn = func(_ context.Context, id int64, _ *models.UpdateProduct) (*models.Product, error) {
					return &models.Product{ID: id, Name: "Updated"}, nil
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

			c, w := newTestContext(http.MethodPut, "/products/"+tt.idParam, tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, tt.callerRole)
			h.UpdateProduct(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				resp := decodeBodyMap(t, w)
				assert.Equal(t, tt.errMessage, resp["error"])
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestHandler_DeleteProduct(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		callerID   int64
		callerRole models.Role
		setup      func(svc *MockProductService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			callerID:   1,
			callerRole: models.RoleAdmin,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:       "seller forbidden — not owner",
			idParam:    "1",
			callerID:   10,
			callerRole: models.RoleSeller,
			setup: func(svc *MockProductService) {
				ownerID := int64(99)
				svc.GetByIDFn = func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id, SellerID: &ownerID}, nil
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "cannot delete another seller's product",
		},
		{
			name:       "service error",
			idParam:    "1",
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockProductService) {
				svc.DeleteFn = func(_ context.Context, _ int64) error {
					return apperrors.NotFound("product not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "product not found",
		},
		{
			name:       "success",
			idParam:    "1",
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockProductService) {
				svc.DeleteFn = func(_ context.Context, _ int64) error { return nil }
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

			c, w := newTestContext(http.MethodDelete, "/products/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, tt.callerRole)
			h.DeleteProduct(c)

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
