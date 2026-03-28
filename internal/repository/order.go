package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	FindByID(ctx context.Context, id int64) (*models.Order, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Order, int, error)
	UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create is used by tests (nil-db path in OrderService). Production creates
// orders via the transactional path in OrderService.Create directly.
func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	const q = `
		INSERT INTO orders (user_id, status, total_price, created_at, updated_at)
		VALUES (:user_id, :status, :total_price, :created_at, :updated_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, order)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&order.ID)
	}

	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*models.Order, error) {
	var order models.Order

	const q = `SELECT * FROM orders WHERE id = $1`

	if err := r.db.GetContext(ctx, &order, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("order not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	const qItems = `SELECT * FROM order_items WHERE order_id = $1`

	if err := r.db.SelectContext(ctx, &order.Items, qItems, id); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return &order, nil
}

func (r *orderRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Order, int, error) {
	var orders []*models.Order

	const q = `SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	if err := r.db.SelectContext(ctx, &orders, q, userID, limit, offset); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	var total int
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID)

	return orders, total, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	const q = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("order not found", nil)
	}

	return nil
}
