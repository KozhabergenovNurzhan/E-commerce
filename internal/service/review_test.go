package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReviewService(reviewRepo *testutil.MockReviewRepo, productRepo *testutil.MockProductRepo) *service.ReviewService {
	return service.NewReviewService(reviewRepo, productRepo)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestReviewCreate(t *testing.T) {
	tests := []struct {
		name        string
		productRepo *testutil.MockProductRepo
		reviewRepo  *testutil.MockReviewRepo
		userID      int64
		productID   int64
		req         *models.CreateReview
		wantCode    int
		check       func(t *testing.T, r *models.Review)
	}{
		{
			name: "success",
			productRepo: &testutil.MockProductRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id, IsActive: true}, nil
				},
			},
			reviewRepo: &testutil.MockReviewRepo{
				CreateFn: func(_ context.Context, r *models.Review) error {
					r.ID = 10
					return nil
				},
			},
			userID:    1,
			productID: 5,
			req:       &models.CreateReview{Rating: 5, Comment: "Отлично!"},
			check: func(t *testing.T, r *models.Review) {
				assert.Equal(t, int64(10), r.ID)
				assert.Equal(t, int64(1), r.UserID)
				assert.Equal(t, int64(5), r.ProductID)
				assert.Equal(t, 5, r.Rating)
				assert.Equal(t, "Отлично!", r.Comment)
				assert.False(t, r.CreatedAt.IsZero())
			},
		},
		{
			name: "product not found",
			productRepo: &testutil.MockProductRepo{
				FindByIDFn: func(_ context.Context, _ int64) (*models.Product, error) {
					return nil, apperrors.NotFound("product not found", nil)
				},
			},
			reviewRepo: &testutil.MockReviewRepo{},
			req:        &models.CreateReview{Rating: 4},
			wantCode:   http.StatusNotFound,
		},
		{
			name: "duplicate review",
			productRepo: &testutil.MockProductRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id}, nil
				},
			},
			reviewRepo: &testutil.MockReviewRepo{
				CreateFn: func(_ context.Context, _ *models.Review) error {
					return apperrors.Conflict("you have already reviewed this product", nil)
				},
			},
			req:      &models.CreateReview{Rating: 3},
			wantCode: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newReviewService(tc.reviewRepo, tc.productRepo)
			got, err := svc.Create(context.Background(), tc.userID, tc.productID, tc.req)

			if tc.wantCode != 0 {
				assertCode(t, err, tc.wantCode)
				return
			}
			require.NoError(t, err)
			tc.check(t, got)
		})
	}
}

// ── ListByProduct ─────────────────────────────────────────────────────────────

func TestReviewListByProduct(t *testing.T) {
	tests := []struct {
		name        string
		productRepo *testutil.MockProductRepo
		reviewRepo  *testutil.MockReviewRepo
		productID   int64
		wantCode    int
		check       func(t *testing.T, reviews []*models.Review, total int, rating *models.ProductRating)
	}{
		{
			name: "success",
			productRepo: &testutil.MockProductRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Product, error) {
					return &models.Product{ID: id}, nil
				},
			},
			reviewRepo: &testutil.MockReviewRepo{
				ListByProductFn: func(_ context.Context, productID int64, _, _ int) ([]*models.Review, int, error) {
					return []*models.Review{
						{ID: 1, ProductID: productID, Rating: 5},
						{ID: 2, ProductID: productID, Rating: 4},
					}, 2, nil
				},
				GetRatingFn: func(_ context.Context, _ int64) (*models.ProductRating, error) {
					return &models.ProductRating{Average: 4.5, Count: 2}, nil
				},
			},
			productID: 1,
			check: func(t *testing.T, reviews []*models.Review, total int, rating *models.ProductRating) {
				assert.Len(t, reviews, 2)
				assert.Equal(t, 2, total)
				assert.Equal(t, 4.5, rating.Average)
				assert.Equal(t, 2, rating.Count)
			},
		},
		{
			name: "product not found",
			productRepo: &testutil.MockProductRepo{
				FindByIDFn: func(_ context.Context, _ int64) (*models.Product, error) {
					return nil, apperrors.NotFound("product not found", nil)
				},
			},
			reviewRepo: &testutil.MockReviewRepo{},
			productID:  99,
			wantCode:   http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newReviewService(tc.reviewRepo, tc.productRepo)
			reviews, total, rating, err := svc.ListByProduct(context.Background(), tc.productID, 1, 20)

			if tc.wantCode != 0 {
				assertCode(t, err, tc.wantCode)
				return
			}
			require.NoError(t, err)
			tc.check(t, reviews, total, rating)
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestReviewUpdate(t *testing.T) {
	tests := []struct {
		name       string
		reviewRepo *testutil.MockReviewRepo
		reviewID   int64
		callerID   int64
		req        *models.UpdateReview
		wantCode   int
		check      func(t *testing.T, r *models.Review)
	}{
		{
			name: "success",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Review, error) {
					return &models.Review{ID: id, UserID: 1, Rating: 3, Comment: "Нормально"}, nil
				},
				UpdateFn: func(_ context.Context, _ *models.Review) error { return nil },
			},
			reviewID: 1,
			callerID: 1,
			req:      &models.UpdateReview{Rating: 5, Comment: "Отлично!"},
			check: func(t *testing.T, r *models.Review) {
				assert.Equal(t, 5, r.Rating)
				assert.Equal(t, "Отлично!", r.Comment)
			},
		},
		{
			name: "not owner",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Review, error) {
					return &models.Review{ID: id, UserID: 2}, nil
				},
			},
			reviewID: 1,
			callerID: 99,
			req:      &models.UpdateReview{Rating: 1},
			wantCode: http.StatusForbidden,
		},
		{
			name: "not found",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, _ int64) (*models.Review, error) {
					return nil, apperrors.NotFound("review not found", nil)
				},
			},
			reviewID: 99,
			callerID: 1,
			req:      &models.UpdateReview{Rating: 4},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newReviewService(tc.reviewRepo, &testutil.MockProductRepo{})
			got, err := svc.Update(context.Background(), tc.reviewID, tc.callerID, tc.req)

			if tc.wantCode != 0 {
				assertCode(t, err, tc.wantCode)
				return
			}
			require.NoError(t, err)
			tc.check(t, got)
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestReviewDelete(t *testing.T) {
	tests := []struct {
		name       string
		reviewRepo *testutil.MockReviewRepo
		reviewID   int64
		callerID   int64
		role       models.Role
		wantCode   int
	}{
		{
			name: "owner can delete",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Review, error) {
					return &models.Review{ID: id, UserID: 1}, nil
				},
				DeleteFn: func(_ context.Context, _ int64) error { return nil },
			},
			reviewID: 1,
			callerID: 1,
			role:     models.RoleCustomer,
		},
		{
			name: "admin can delete any review",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Review, error) {
					return &models.Review{ID: id, UserID: 5}, nil
				},
				DeleteFn: func(_ context.Context, _ int64) error { return nil },
			},
			reviewID: 1,
			callerID: 99,
			role:     models.RoleAdmin,
		},
		{
			name: "not owner forbidden",
			reviewRepo: &testutil.MockReviewRepo{
				FindByIDFn: func(_ context.Context, id int64) (*models.Review, error) {
					return &models.Review{ID: id, UserID: 5}, nil
				},
			},
			reviewID: 1,
			callerID: 99,
			role:     models.RoleCustomer,
			wantCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newReviewService(tc.reviewRepo, &testutil.MockProductRepo{})
			err := svc.Delete(context.Background(), tc.reviewID, tc.callerID, tc.role)

			if tc.wantCode != 0 {
				assertCode(t, err, tc.wantCode)
				return
			}
			require.NoError(t, err)
		})
	}
}
