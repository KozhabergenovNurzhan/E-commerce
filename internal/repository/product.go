package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	FindByID(ctx context.Context, id int64) (*models.Product, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error)
	ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
	CreateCategory(ctx context.Context, c *models.Category) error
	UpdateCategory(ctx context.Context, c *models.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *models.Product) error {
	const q = `
		INSERT INTO products (category_id, seller_id, name, description, price, stock, image_url, is_active, created_at, updated_at)
		VALUES (:category_id, :seller_id, :name, :description, :price, :stock, :image_url, :is_active, :created_at, :updated_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&p.ID)
	}

	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	var p models.Product

	const q = `SELECT * FROM products WHERE id = $1 AND is_active = true`

	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("product not found", nil)
		}
		return nil, apperrors.Internal("internal server error", err)
	}

	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, p *models.Product) error {
	const q = `
		UPDATE products
		SET name = :name, description = :description, price = :price,
		    stock = :stock, image_url = :image_url, updated_at = :updated_at
		WHERE id = :id AND is_active = true`

	result, err := r.db.NamedExecContext(ctx, q, p)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("product not found", nil)
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	const q = `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1 AND is_active = true`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("product not found", nil)
	}

	return nil
}

func (r *productRepository) List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
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
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	var products []*models.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	return products, total, nil
}

func (r *productRepository) ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error) {
	query := `SELECT * FROM products WHERE is_active = true AND seller_id = $1`
	count := `SELECT COUNT(*) FROM products WHERE is_active = true AND seller_id = $1`
	args := []interface{}{sellerID}
	i := 2

	if f.Search != "" {
		clause := fmt.Sprintf(" AND name ILIKE $%d", i)
		query += clause
		count += clause
		args = append(args, "%"+f.Search+"%")
		i++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, count, args...); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	var products []*models.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, apperrors.Internal("internal server error", err)
	}

	return products, total, nil
}

func (r *productRepository) ListCategories(ctx context.Context) ([]*models.Category, error) {
	var cats []*models.Category

	if err := r.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY name`); err != nil {
		return nil, apperrors.Internal("internal server error", err)
	}

	return cats, nil
}

func (r *productRepository) CreateCategory(ctx context.Context, c *models.Category) error {
	const q = `
		INSERT INTO categories (name, slug, created_at)
		VALUES (:name, :slug, :created_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, q, c)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&c.ID)
	}

	return nil
}

func (r *productRepository) UpdateCategory(ctx context.Context, c *models.Category) error {
	const q = `UPDATE categories SET name = :name, slug = :slug WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, q, c)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("category not found", nil)
	}

	return nil
}

func (r *productRepository) DeleteCategory(ctx context.Context, id int64) error {
	const q = `DELETE FROM categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return apperrors.Internal("internal server error", err)
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.NotFound("category not found", nil)
	}

	return nil
}
