package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"
)

type ReviewRepository interface {
	Create(ctx context.Context, r *models.Review) error
	FindByID(ctx context.Context, id int64) (*models.Review, error)
	ListByProduct(ctx context.Context, productID int64, limit, offset int) ([]*models.Review, int, error)
	Update(ctx context.Context, r *models.Review) error
	Delete(ctx context.Context, id int64) error
	GetRating(ctx context.Context, productID int64) (*models.ProductRating, error)
}

type reviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, rev *models.Review) error {
	const q = `
		INSERT INTO reviews (product_id, user_id, rating, comment, created_at, updated_at)
		VALUES (:product_id, :user_id, :rating, :comment, :created_at, :updated_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, rev)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return apperrors.Conflict("you have already reviewed this product", nil)
		}
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&rev.ID)
	}

	return nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id int64) (*models.Review, error) {
	var rev models.Review

	if err := r.db.GetContext(ctx, &rev, `SELECT * FROM reviews WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("review not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &rev, nil
}

func (r *reviewRepository) ListByProduct(ctx context.Context, productID int64, limit, offset int) ([]*models.Review, int, error) {
	var reviews []*models.Review

	const q = `SELECT * FROM reviews WHERE product_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	if err := r.db.SelectContext(ctx, &reviews, q, productID, limit, offset); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM reviews WHERE product_id = $1`, productID); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) Update(ctx context.Context, rev *models.Review) error {
	const q = `UPDATE reviews SET rating = :rating, comment = :comment, updated_at = :updated_at WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, q, rev)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("review not found", nil)
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("review not found", nil)
	}

	return nil
}

func (r *reviewRepository) GetRating(ctx context.Context, productID int64) (*models.ProductRating, error) {
	var rating models.ProductRating

	const q = `SELECT COALESCE(AVG(rating), 0) AS average, COUNT(*) AS count FROM reviews WHERE product_id = $1`

	if err := r.db.GetContext(ctx, &rating, q, productID); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return &rating, nil
}
