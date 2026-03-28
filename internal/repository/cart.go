package repository

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type CartRepository interface {
	Upsert(ctx context.Context, item *models.CartItemRecord) error
	FindByUserID(ctx context.Context, userID int64) ([]*models.CartItemRecord, error)
	Delete(ctx context.Context, userID, productID int64) error
	Clear(ctx context.Context, userID int64) error
}

type cartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Upsert(ctx context.Context, item *models.CartItemRecord) error {
	const q = `
		INSERT INTO cart_items (user_id, product_id, quantity, created_at, updated_at)
		VALUES (:user_id, :product_id, :quantity, :created_at, :updated_at)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = :quantity, updated_at = :updated_at
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, item)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&item.ID)
	}

	return nil
}

func (r *cartRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.CartItemRecord, error) {
	var items []*models.CartItemRecord

	const q = `SELECT * FROM cart_items WHERE user_id = $1 ORDER BY created_at ASC`

	if err := r.db.SelectContext(ctx, &items, q, userID); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return items, nil
}

func (r *cartRepository) Delete(ctx context.Context, userID, productID int64) error {
	const q = `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`

	if _, err := r.db.ExecContext(ctx, q, userID, productID); err != nil {
		return apperrors.Internal("internal server error", err)

	}

	return nil
}

func (r *cartRepository) Clear(ctx context.Context, userID int64) error {
	const q = `DELETE FROM cart_items WHERE user_id = $1`

	if _, err := r.db.ExecContext(ctx, q, userID); err != nil {
		return apperrors.Internal("internal server error", err)

	}

	return nil
}
