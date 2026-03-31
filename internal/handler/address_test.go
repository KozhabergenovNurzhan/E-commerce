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

func TestHandler_CreateAddress(t *testing.T) {
	validBody := `{"full_name":"John Doe","phone":"+77001234567","country":"Kazakhstan",
					"city":"Almaty","street":"Abay 1","postal_code":"050000"}`

	tests := []struct {
		name       string
		body       string
		callerID   int64
		setup      func(svc *MockAddressService)
		errCode    int
		errMessage string
	}{
		{
			name:     "invalid body",
			body:     `{"full_name":`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "service error",
			body:     validBody,
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.CreateFn = func(_ context.Context, _ int64, _ *models.CreateAddress) (*models.Address, error) {
					return nil, apperrors.Internal("internal server error", nil)
				}
			},
			errCode:    http.StatusInternalServerError,
			errMessage: "internal server error",
		},
		{
			name:     "success",
			body:     validBody,
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.CreateFn = func(_ context.Context, userID int64, req *models.CreateAddress) (*models.Address, error) {
					return &models.Address{ID: 5, UserID: userID, City: req.City}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrSvc := &MockAddressService{}
			if tt.setup != nil {
				tt.setup(addrSvc)
			}
			h := newTestHandler(&service.Services{Address: addrSvc})

			c, w := newTestContext(http.MethodPost, "/addresses", tt.body)
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.CreateAddress(c)

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

func TestHandler_UpdateAddress(t *testing.T) {
	validBody := `{"full_name":"Jane Doe","phone":"+77009876543","country":"Kazakhstan","city":"Astana","street":"Kerey 5","postal_code":"010000"}`

	tests := []struct {
		name       string
		idParam    string
		body       string
		callerID   int64
		setup      func(svc *MockAddressService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			body:       validBody,
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid address id",
		},
		{
			name:     "not owner",
			idParam:  "1",
			body:     validBody,
			callerID: 99,
			setup: func(svc *MockAddressService) {
				svc.UpdateFn = func(_ context.Context, _, _ int64, _ *models.UpdateAddress) (*models.Address, error) {
					return nil, apperrors.Forbidden("forbidden", nil)
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "forbidden",
		},
		{
			name:     "success",
			idParam:  "1",
			body:     validBody,
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.UpdateFn = func(_ context.Context, addressID, userID int64, req *models.UpdateAddress) (*models.Address, error) {
					return &models.Address{ID: addressID, UserID: userID, City: req.City}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrSvc := &MockAddressService{}
			if tt.setup != nil {
				tt.setup(addrSvc)
			}
			h := newTestHandler(&service.Services{Address: addrSvc})

			c, w := newTestContext(http.MethodPut, "/addresses/"+tt.idParam, tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.UpdateAddress(c)

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

func TestHandler_DeleteAddress(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		callerID   int64
		setup      func(svc *MockAddressService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid address id",
		},
		{
			name:     "not owner",
			idParam:  "1",
			callerID: 99,
			setup: func(svc *MockAddressService) {
				svc.DeleteFn = func(_ context.Context, _, _ int64) error {
					return apperrors.Forbidden("forbidden", nil)
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "forbidden",
		},
		{
			name:     "not found",
			idParam:  "99",
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.DeleteFn = func(_ context.Context, _, _ int64) error {
					return apperrors.NotFound("address not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "address not found",
		},
		{
			name:     "success",
			idParam:  "1",
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.DeleteFn = func(_ context.Context, addressID, userID int64) error {
					assert.Equal(t, int64(1), addressID)
					assert.Equal(t, int64(1), userID)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrSvc := &MockAddressService{}
			if tt.setup != nil {
				tt.setup(addrSvc)
			}
			h := newTestHandler(&service.Services{Address: addrSvc})

			c, w := newTestContext(http.MethodDelete, "/addresses/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.DeleteAddress(c)

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

func TestHandler_SetDefaultAddress(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		callerID   int64
		setup      func(svc *MockAddressService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "xyz",
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid address id",
		},
		{
			name:     "not owner",
			idParam:  "1",
			callerID: 99,
			setup: func(svc *MockAddressService) {
				svc.SetDefaultFn = func(_ context.Context, _, _ int64) (*models.Address, error) {
					return nil, apperrors.Forbidden("forbidden", nil)
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "forbidden",
		},
		{
			name:     "success",
			idParam:  "1",
			callerID: 1,
			setup: func(svc *MockAddressService) {
				svc.SetDefaultFn = func(_ context.Context, addressID, userID int64) (*models.Address, error) {
					return &models.Address{ID: addressID, UserID: userID, IsDefault: true}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrSvc := &MockAddressService{}
			if tt.setup != nil {
				tt.setup(addrSvc)
			}
			h := newTestHandler(&service.Services{Address: addrSvc})

			c, w := newTestContext(http.MethodPatch, "/addresses/"+tt.idParam+"/default", "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.SetDefaultAddress(c)

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
