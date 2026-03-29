package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type ReviewService struct {
	repo        repository.ReviewRepository
	productRepo repository.ProductRepository
}

func NewReviewService(repo repository.ReviewRepository, productRepo repository.ProductRepository) *ReviewService {
	return &ReviewService{repo: repo, productRepo: productRepo}
}

func (s *ReviewService) Create(ctx context.Context, userID, productID int64, req *models.CreateReview) (*models.Review, error) {
	if _, err := s.productRepo.FindByID(ctx, productID); err != nil {
		return nil, err
	}

	now := utils.Now()
	review := &models.Review{
		ProductID: productID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) ListByProduct(ctx context.Context, productID int64, page, limit int) ([]*models.Review, int, *models.ProductRating, error) {
	if _, err := s.productRepo.FindByID(ctx, productID); err != nil {
		return nil, 0, nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	reviews, total, err := s.repo.ListByProduct(ctx, productID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, nil, err
	}

	rating, err := s.repo.GetRating(ctx, productID)
	if err != nil {
		return nil, 0, nil, err
	}

	return reviews, total, rating, nil
}

func (s *ReviewService) Update(ctx context.Context, reviewID, userID int64, req *models.UpdateReview) (*models.Review, error) {
	review, err := s.repo.FindByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.UserID != userID {
		return nil, apperrors.Forbidden("cannot edit another user's review", nil)
	}

	review.Rating = req.Rating
	review.Comment = req.Comment
	review.UpdatedAt = utils.Now()

	if err := s.repo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, reviewID, callerID int64, callerRole models.Role) error {
	review, err := s.repo.FindByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if callerRole != models.RoleAdmin && review.UserID != callerID {
		return apperrors.Forbidden("cannot delete another user's review", nil)
	}

	return s.repo.Delete(ctx, reviewID)
}
