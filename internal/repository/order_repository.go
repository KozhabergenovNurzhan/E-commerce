package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.ErrInternal
	}
	defer tx.Rollback()

	const qOrder = `
		INSERT INTO orders (id, user_id, status, total_price, created_at, updated_at)
		VALUES (:id, :user_id, :status, :total_price, :created_at, :updated_at)`
	if _, err := tx.NamedExecContext(ctx, qOrder, order); err != nil {
		return apperrors.ErrInternal
	}

	const qItem = `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price)
		VALUES (:id, :order_id, :product_id, :quantity, :unit_price)`
	for _, item := range order.Items {
		if _, err := tx.NamedExecContext(ctx, qItem, item); err != nil {
			return apperrors.ErrInternal
		}
	}

	return tx.Commit()
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	const q = `SELECT * FROM orders WHERE id = $1`
	if err := r.db.GetContext(ctx, &order, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.ErrInternal
	}

	const qItems = `SELECT * FROM order_items WHERE order_id = $1`
	if err := r.db.SelectContext(ctx, &order.Items, qItems, id); err != nil {
		return nil, apperrors.ErrInternal
	}
	return &order, nil
}

func (r *orderRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error) {
	var orders []*domain.Order
	const q = `SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &orders, q, userID, limit, offset); err != nil {
		return nil, 0, apperrors.ErrInternal
	}
	var total int
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID)
	return orders, total, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	const q = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return apperrors.ErrInternal
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
