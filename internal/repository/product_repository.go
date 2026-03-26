package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/apperrors"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	FindByID(ctx context.Context, id int64) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error)
	ListCategories(ctx context.Context) ([]*domain.Category, error)
}

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	const q = `
		INSERT INTO products (category_id, name, description, price, stock, image_url, is_active, created_at, updated_at)
		VALUES (:category_id, :name, :description, :price, :stock, :image_url, :is_active, :created_at, :updated_at)
		RETURNING id`
	rows, err := r.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		return apperrors.ErrInternal
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&p.ID)
	}
	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product
	const q = `SELECT * FROM products WHERE id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.ErrInternal
	}
	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	const q = `
		UPDATE products
		SET name = :name, description = :description, price = :price,
		    stock = :stock, image_url = :image_url, updated_at = :updated_at
		WHERE id = :id AND is_active = true`
	result, err := r.db.NamedExecContext(ctx, q, p)
	if err != nil {
		return apperrors.ErrInternal
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	const q = `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1 AND is_active = true`
	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.ErrInternal
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *productRepository) List(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error) {
	query := `SELECT * FROM products WHERE is_active = true`
	count := `SELECT COUNT(*) FROM products WHERE is_active = true`
	args := []interface{}{}
	i := 1

	if f.Search != "" {
		clause := fmt.Sprintf(" AND name ILIKE $%d", i)
		query += clause
		count += clause
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.CategoryID != nil {
		clause := fmt.Sprintf(" AND category_id = $%d", i)
		query += clause
		count += clause
		args = append(args, *f.CategoryID)
		i++
	}
	if f.MinPrice != nil {
		clause := fmt.Sprintf(" AND price >= $%d", i)
		query += clause
		count += clause
		args = append(args, *f.MinPrice)
		i++
	}
	if f.MaxPrice != nil {
		clause := fmt.Sprintf(" AND price <= $%d", i)
		query += clause
		count += clause
		args = append(args, *f.MaxPrice)
		i++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, count, args...); err != nil {
		return nil, 0, apperrors.ErrInternal
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, apperrors.ErrInternal
	}
	return products, total, nil
}

func (r *productRepository) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	var cats []*domain.Category
	if err := r.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY name`); err != nil {
		return nil, apperrors.ErrInternal
	}
	return cats, nil
}
