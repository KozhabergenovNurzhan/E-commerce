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

func TestHandler_ListReviews(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setup      func(svc *MockReviewService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid product id",
			idParam:    "abc",
			errCode:    http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:    "product not found",
			idParam: "99",
			setup: func(svc *MockReviewService) {
				svc.ListByProductFn = func(_ context.Context, _ int64, _, _ int) ([]*models.Review, int, *models.ProductRating, error) {
					return nil, 0, nil, apperrors.NotFound("product not found", nil)
				}
			},
			errCode:    http.StatusNotFound,
			errMessage: "product not found",
		},
		{
			name:    "success",
			idParam: "1",
			setup: func(svc *MockReviewService) {
				svc.ListByProductFn = func(_ context.Context, productID int64, _, _ int) ([]*models.Review, int, *models.ProductRating, error) {
					return []*models.Review{
						{ID: 1, ProductID: productID, Rating: 5, Comment: "Great!"},
					}, 1, &models.ProductRating{Average: 5.0, Count: 1}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewSvc := &MockReviewService{}
			if tt.setup != nil {
				tt.setup(reviewSvc)
			}
			h := newTestHandler(&service.Services{Review: reviewSvc})

			c, w := newTestContext(http.MethodGet, "/products/"+tt.idParam+"/reviews", "")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			h.ListReviews(c)

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

func TestHandler_CreateReview(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		body       string
		callerID   int64
		setup      func(svc *MockReviewService)
		errCode    int
		errMessage string
	}{
		{
			name:     "invalid product id",
			idParam:  "abc",
			body:     `{"rating":5}`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
			errMessage: "invalid product id",
		},
		{
			name:     "invalid body",
			idParam:  "1",
			body:     `{"rating":`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "duplicate review",
			idParam:  "1",
			body:     `{"rating":5,"comment":"Great!"}`,
			callerID: 1,
			setup: func(svc *MockReviewService) {
				svc.CreateFn = func(_ context.Context, _, _ int64, _ *models.CreateReview) (*models.Review, error) {
					return nil, apperrors.Conflict("you have already reviewed this product", nil)
				}
			},
			errCode:    http.StatusConflict,
			errMessage: "you have already reviewed this product",
		},
		{
			name:     "success",
			idParam:  "1",
			body:     `{"rating":5,"comment":"Great!"}`,
			callerID: 1,
			setup: func(svc *MockReviewService) {
				svc.CreateFn = func(_ context.Context, userID, productID int64, req *models.CreateReview) (*models.Review, error) {
					return &models.Review{
						ID:        10,
						UserID:    userID,
						ProductID: productID,
						Rating:    req.Rating,
						Comment:   req.Comment,
					}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewSvc := &MockReviewService{}
			if tt.setup != nil {
				tt.setup(reviewSvc)
			}
			h := newTestHandler(&service.Services{Review: reviewSvc})

			c, w := newTestContext(http.MethodPost, "/products/"+tt.idParam+"/reviews", tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.CreateReview(c)

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

func TestHandler_UpdateReview(t *testing.T) {
	tests := []struct {
		name       string
		reviewID   string
		body       string
		callerID   int64
		setup      func(svc *MockReviewService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid review id",
			reviewID:   "abc",
			body:       `{"rating":4}`,
			callerID:   1,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid review id",
		},
		{
			name:     "invalid body",
			reviewID: "1",
			body:     `{"rating":`,
			callerID: 1,
			errCode:  http.StatusBadRequest,
		},
		{
			name:     "not owner",
			reviewID: "1",
			body:     `{"rating":4,"comment":"OK"}`,
			callerID: 99,
			setup: func(svc *MockReviewService) {
				svc.UpdateFn = func(_ context.Context, _, _ int64, _ *models.UpdateReview) (*models.Review, error) {
					return nil, apperrors.Forbidden("forbidden", nil)
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "forbidden",
		},
		{
			name:     "success",
			reviewID: "1",
			body:     `{"rating":5,"comment":"Excellent!"}`,
			callerID: 1,
			setup: func(svc *MockReviewService) {
				svc.UpdateFn = func(_ context.Context, reviewID, _ int64, req *models.UpdateReview) (*models.Review, error) {
					return &models.Review{ID: reviewID, Rating: req.Rating, Comment: req.Comment}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewSvc := &MockReviewService{}
			if tt.setup != nil {
				tt.setup(reviewSvc)
			}
			h := newTestHandler(&service.Services{Review: reviewSvc})

			c, w := newTestContext(http.MethodPut, "/products/1/reviews/"+tt.reviewID, tt.body)
			c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "reviewId", Value: tt.reviewID}}
			setAuth(c, tt.callerID, models.RoleCustomer)
			h.UpdateReview(c)

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

func TestHandler_DeleteReview(t *testing.T) {
	tests := []struct {
		name       string
		reviewID   string
		callerID   int64
		callerRole models.Role
		setup      func(svc *MockReviewService)
		errCode    int
		errMessage string
	}{
		{
			name:       "invalid review id",
			reviewID:   "abc",
			callerID:   1,
			callerRole: models.RoleCustomer,
			errCode:    http.StatusBadRequest,
			errMessage: "invalid review id",
		},
		{
			name:       "not owner",
			reviewID:   "1",
			callerID:   99,
			callerRole: models.RoleCustomer,
			setup: func(svc *MockReviewService) {
				svc.DeleteFn = func(_ context.Context, _, _ int64, _ models.Role) error {
					return apperrors.Forbidden("forbidden", nil)
				}
			},
			errCode:    http.StatusForbidden,
			errMessage: "forbidden",
		},
		{
			name:       "success as owner",
			reviewID:   "1",
			callerID:   1,
			callerRole: models.RoleCustomer,
			setup: func(svc *MockReviewService) {
				svc.DeleteFn = func(_ context.Context, reviewID, callerID int64, role models.Role) error {
					assert.Equal(t, int64(1), reviewID)
					assert.Equal(t, int64(1), callerID)
					assert.Equal(t, models.RoleCustomer, role)
					return nil
				}
			},
		},
		{
			name:       "success as admin",
			reviewID:   "5",
			callerID:   99,
			callerRole: models.RoleAdmin,
			setup: func(svc *MockReviewService) {
				svc.DeleteFn = func(_ context.Context, _, _ int64, role models.Role) error {
					assert.Equal(t, models.RoleAdmin, role)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewSvc := &MockReviewService{}
			if tt.setup != nil {
				tt.setup(reviewSvc)
			}
			h := newTestHandler(&service.Services{Review: reviewSvc})

			c, w := newTestContext(http.MethodDelete, "/products/1/reviews/"+tt.reviewID, "")
			c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "reviewId", Value: tt.reviewID}}
			setAuth(c, tt.callerID, tt.callerRole)
			h.DeleteReview(c)

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
