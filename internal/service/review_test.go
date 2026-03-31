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

// ── Create ────────────────────────────────────────────────────────────────────

func TestReviewService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int64
		productID  int64
		req        *models.CreateReview
		setup      func(reviewRepo *MockReviewRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, r *models.Review)
	}{
		{
			name:      "success",
			userID:    1,
			productID: 5,
			req:       &models.CreateReview{Rating: 5, Comment: "Excellent!"},
			setup: func(reviewRepo *MockReviewRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(5)).
					Return(&models.Product{ID: 5}, nil).Once()
				reviewRepo.On("Create", ctx, mock.AnythingOfType("*models.Review")).
					Run(func(args mock.Arguments) {
						args.Get(1).(*models.Review).ID = 10
					}).
					Return(nil).Once()
			},
			check: func(t *testing.T, r *models.Review) {
				assert.Equal(t, int64(10), r.ID)
				assert.Equal(t, int64(1), r.UserID)
				assert.Equal(t, int64(5), r.ProductID)
				assert.Equal(t, 5, r.Rating)
				assert.Equal(t, "Excellent!", r.Comment)
				assert.False(t, r.CreatedAt.IsZero())
			},
		},
		{
			name: "product not found",
			req:  &models.CreateReview{Rating: 4},
			setup: func(_ *MockReviewRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(0)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
		{
			name:      "duplicate review",
			productID: 5,
			req:       &models.CreateReview{Rating: 3},
			setup: func(reviewRepo *MockReviewRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(5)).
					Return(&models.Product{ID: 5}, nil).Once()
				reviewRepo.On("Create", ctx, mock.AnythingOfType("*models.Review")).
					Return(apperrors.Conflict("you have already reviewed this product", nil)).Once()
			},
			errCode: http.StatusConflict,
			errMsg:  "you have already reviewed this product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewRepo := new(MockReviewRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(reviewRepo, productRepo)
			}

			svc := service.NewReviewService(reviewRepo, productRepo)
			got, err := svc.Create(ctx, tt.userID, tt.productID, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, got)
			}

			reviewRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

// ── ListByProduct ─────────────────────────────────────────────────────────────

func TestReviewService_ListByProduct(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		productID  int64
		setup      func(reviewRepo *MockReviewRepo, productRepo *MockProductRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, reviews []*models.Review, total int, rating *models.ProductRating)
	}{
		{
			name:      "success",
			productID: 1,
			setup: func(reviewRepo *MockReviewRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(1)).
					Return(&models.Product{ID: 1}, nil).Once()
				reviewRepo.On("ListByProduct", ctx, int64(1), 20, 0).
					Return([]*models.Review{
						{ID: 1, ProductID: 1, Rating: 5},
						{ID: 2, ProductID: 1, Rating: 4},
					}, 2, nil).Once()
				reviewRepo.On("GetRating", ctx, int64(1)).
					Return(&models.ProductRating{Average: 4.5, Count: 2}, nil).Once()
			},
			check: func(t *testing.T, reviews []*models.Review, total int, rating *models.ProductRating) {
				assert.Len(t, reviews, 2)
				assert.Equal(t, 2, total)
				assert.Equal(t, 4.5, rating.Average)
				assert.Equal(t, 2, rating.Count)
			},
		},
		{
			name:      "product not found",
			productID: 99,
			setup: func(_ *MockReviewRepo, productRepo *MockProductRepo) {
				productRepo.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("product not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewRepo := new(MockReviewRepo)
			productRepo := new(MockProductRepo)
			if tt.setup != nil {
				tt.setup(reviewRepo, productRepo)
			}

			svc := service.NewReviewService(reviewRepo, productRepo)
			reviews, total, rating, err := svc.ListByProduct(ctx, tt.productID, 1, 20)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, reviews, total, rating)
			}

			reviewRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestReviewService_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		reviewID   int64
		callerID   int64
		req        *models.UpdateReview
		setup      func(r *MockReviewRepo)
		errCode    int
		errMsg     string
		errWrapped error
		check      func(t *testing.T, r *models.Review)
	}{
		{
			name:     "success",
			reviewID: 1,
			callerID: 1,
			req:      &models.UpdateReview{Rating: 5, Comment: "Excellent!"},
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Review{ID: 1, UserID: 1, Rating: 3, Comment: "OK"}, nil).Once()
				r.On("Update", ctx, mock.AnythingOfType("*models.Review")).
					Return(nil).Once()
			},
			check: func(t *testing.T, r *models.Review) {
				assert.Equal(t, 5, r.Rating)
				assert.Equal(t, "Excellent!", r.Comment)
			},
		},
		{
			name:     "not owner",
			reviewID: 1,
			callerID: 99,
			req:      &models.UpdateReview{Rating: 1},
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Review{ID: 1, UserID: 2}, nil).Once()
			},
			errCode: http.StatusForbidden,
			errMsg:  "cannot edit another user's review",
		},
		{
			name:     "review not found",
			reviewID: 99,
			callerID: 1,
			req:      &models.UpdateReview{Rating: 4},
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(99)).
					Return(nil, apperrors.NotFound("review not found", nil)).Once()
			},
			errCode: http.StatusNotFound,
			errMsg:  "review not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewRepo := new(MockReviewRepo)
			if tt.setup != nil {
				tt.setup(reviewRepo)
			}

			svc := service.NewReviewService(reviewRepo, new(MockProductRepo))
			got, err := svc.Update(ctx, tt.reviewID, tt.callerID, tt.req)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
				tt.check(t, got)
			}

			reviewRepo.AssertExpectations(t)
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestReviewService_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		reviewID   int64
		callerID   int64
		role       models.Role
		setup      func(r *MockReviewRepo)
		errCode    int
		errMsg     string
		errWrapped error
	}{
		{
			name:     "owner can delete",
			reviewID: 1,
			callerID: 1,
			role:     models.RoleCustomer,
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Review{ID: 1, UserID: 1}, nil).Once()
				r.On("Delete", ctx, int64(1)).Return(nil).Once()
			},
		},
		{
			name:     "admin can delete any review",
			reviewID: 1,
			callerID: 99,
			role:     models.RoleAdmin,
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Review{ID: 1, UserID: 5}, nil).Once()
				r.On("Delete", ctx, int64(1)).Return(nil).Once()
			},
		},
		{
			name:     "not owner forbidden",
			reviewID: 1,
			callerID: 99,
			role:     models.RoleCustomer,
			setup: func(r *MockReviewRepo) {
				r.On("FindByID", ctx, int64(1)).
					Return(&models.Review{ID: 1, UserID: 5}, nil).Once()
			},
			errCode: http.StatusForbidden,
			errMsg:  "cannot delete another user's review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewRepo := new(MockReviewRepo)
			if tt.setup != nil {
				tt.setup(reviewRepo)
			}

			svc := service.NewReviewService(reviewRepo, new(MockProductRepo))
			err := svc.Delete(ctx, tt.reviewID, tt.callerID, tt.role)

			if tt.errCode != 0 {
				require.Error(t, err)
				assertAppErr(t, err, tt.errCode, tt.errMsg, tt.errWrapped)
			} else {
				require.NoError(t, err)
			}

			reviewRepo.AssertExpectations(t)
		})
	}
}
