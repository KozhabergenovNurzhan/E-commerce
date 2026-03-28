package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/utils"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

type OrderService struct {
	db          *sqlx.DB
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderService(db *sqlx.DB, orderRepo repository.OrderRepository, productRepo repository.ProductRepository) *OrderService {
	return &OrderService{db: db, orderRepo: orderRepo, productRepo: productRepo}
}

func (s *OrderService) Create(ctx context.Context, userID int64, req *models.CreateOrder) (*models.Order, error) {
	var total float64
	items := make([]models.OrderItem, 0, len(req.Items))

	for _, r := range req.Items {
		product, err := s.productRepo.FindByID(ctx, r.ProductID)
		if err != nil {
			return nil, err
		}
		if product.Stock < r.Quantity {
			return nil, apperrors.BadRequest("insufficient stock", nil)
		}
		total += product.Price * float64(r.Quantity)

		items = append(items, models.OrderItem{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
			UnitPrice: product.Price,
		})
	}

	now := utils.Now()
	order := &models.Order{
		UserID:     userID,
		Status:     models.OrderStatusPending,
		TotalPrice: total,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      items,
	}

	if s.db == nil {
		// test path: no real DB, delegate to repo mock
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return nil, err
		}
		return order, nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}
	defer tx.Rollback()

	const qOrder = `
		INSERT INTO orders (user_id, status, total_price, created_at, updated_at)
		VALUES (:user_id, :status, :total_price, :created_at, :updated_at)
		RETURNING id`
	rows, err := sqlx.NamedQueryContext(ctx, tx, qOrder, order)
	if err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&order.ID)
	}
	rows.Close()

	const qItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES (:order_id, :product_id, :quantity, :unit_price)
		RETURNING id`
	for i := range order.Items {
		order.Items[i].OrderID = order.ID

		rows, err := sqlx.NamedQueryContext(ctx, tx, qItem, order.Items[i])
		if err != nil {
			return nil, apperrors.Internal("internal server error", err)
		}
		if rows.Next() {
			rows.Scan(&order.Items[i].ID)
		}
		rows.Close()
	}

	const qStock = `UPDATE products SET stock = stock - $1 WHERE id = $2`
	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, qStock, item.Quantity, item.ProductID); err != nil {
			return nil, apperrors.Internal("internal server error", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *OrderService) ListByUser(ctx context.Context, userID int64, page, limit int) ([]*models.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.orderRepo.ListByUser(ctx, userID, limit, (page-1)*limit)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	if _, err := s.orderRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.orderRepo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) Cancel(ctx context.Context, id int64) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != models.OrderStatusPending {
		return apperrors.BadRequest("only pending orders can be cancelled", nil)
	}
	return s.orderRepo.UpdateStatus(ctx, id, models.OrderStatusCancelled)
}
