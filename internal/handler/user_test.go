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

func TestHandler_GetUserByID(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setup      func(userSvc *MockUserService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			errCode:    http.StatusBadRequest,
			errMessage: "invalid user id",
		},
		{
			name:    "not found",
			idParam: "99",
			setup: func(userSvc *MockUserService) {
				userSvc.GetByIDFn = func(_ context.Context, _ int64) (*models.User, error) {
					return nil, apperrors.NotFound("user not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "user not found",
		},
		{
			name:    "success",
			idParam: "1",
			setup: func(userSvc *MockUserService) {
				userSvc.GetByIDFn = func(_ context.Context, id int64) (*models.User, error) {
					return &models.User{ID: id, Email: "john@example.com"}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &MockUserService{}
			if tt.setup != nil {
				tt.setup(userSvc)
			}
			h := newTestHandler(&service.Services{User: userSvc})

			c, w := newTestContext(http.MethodGet, "/users/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.GetUserByID(c)

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

func TestHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		body       string
		callerID   int64
		callerRole models.Role
		setup      func(userSvc *MockUserService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			body:       `{"first_name":"John","last_name":"Doe"}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid user id",
		},
		{
			name:       "forbidden — not owner and not admin",
			idParam:    "2",
			body:       `{"first_name":"John","last_name":"Doe"}`,
			callerID:   1,
			callerRole: models.RoleCustomer,
			errCode:    http.StatusForbidden,
			errMessage: "cannot update another user's profile",
		},
		{
			name:       "invalid body",
			idParam:    "1",
			body:       `{"first_name":`,
			callerID:   1,
			callerRole: models.RoleCustomer,
			errCode:    http.StatusBadRequest,
		},
		{
			name:       "service error",
			idParam:    "1",
			body:       `{"first_name":"John","last_name":"Doe"}`,
			callerID:   1,
			callerRole: models.RoleCustomer,
			setup: func(userSvc *MockUserService) {
				userSvc.UpdateFn = func(_ context.Context, _ int64, _, _ string) (*models.User, error) {
					return nil, apperrors.NotFound("user not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "user not found",
		},
		{
			name:       "success as owner",
			idParam:    "1",
			body:       `{"first_name":"John","last_name":"Doe"}`,
			callerID:   1,
			callerRole: models.RoleCustomer,
			setup: func(userSvc *MockUserService) {
				userSvc.UpdateFn = func(_ context.Context, id int64, fn, ln string) (*models.User, error) {
					return &models.User{ID: id, FirstName: fn, LastName: ln}, nil
				}
			},
		},
		{
			name:       "success as admin updating another user",
			idParam:    "5",
			body:       `{"first_name":"Jane","last_name":"Smith"}`,
			callerID:   1,
			callerRole: models.RoleAdmin,
			setup: func(userSvc *MockUserService) {
				userSvc.UpdateFn = func(_ context.Context, id int64, fn, ln string) (*models.User, error) {
					return &models.User{ID: id, FirstName: fn, LastName: ln}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &MockUserService{}
			if tt.setup != nil {
				tt.setup(userSvc)
			}
			h := newTestHandler(&service.Services{User: userSvc})

			c, w := newTestContext(http.MethodPut, "/users/"+tt.idParam, tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, tt.callerRole)
			h.UpdateUser(c)

			if tt.errCode != 0 {
				require.Equal(t, tt.errCode, w.Code)
				if tt.errMessage != "" {
					resp := decodeBodyMap(t, w)
					assert.Equal(t, tt.errMessage, resp["error"])
				}
				return
			}
			require.Equal(t, http.StatusOK, w.Code)
			resp := decodeBodyMap(t, w)
			assert.True(t, resp["success"].(bool))
		})
	}
}

func TestHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setup      func(userSvc *MockUserService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid id",
			idParam:    "abc",
			errCode:    http.StatusBadRequest,
			errMessage: "invalid user id",
		},
		{
			name:    "not found",
			idParam: "99",
			setup: func(userSvc *MockUserService) {
				userSvc.DeleteFn = func(_ context.Context, _ int64) error {
					return apperrors.NotFound("user not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "user not found",
		},
		{
			name:    "success",
			idParam: "1",
			setup: func(userSvc *MockUserService) {
				userSvc.DeleteFn = func(_ context.Context, _ int64) error { return nil }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &MockUserService{}
			if tt.setup != nil {
				tt.setup(userSvc)
			}
			h := newTestHandler(&service.Services{User: userSvc})

			c, w := newTestContext(http.MethodDelete, "/users/"+tt.idParam, "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.DeleteUser(c)

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
