package domain

import "time"

type Product struct {
	ID          int64     `db:"id"`
	CategoryID  int64     `db:"category_id"`
	SellerID    *int64    `db:"seller_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	ImageURL    string    `db:"image_url"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	CreatedAt time.Time `db:"created_at"`
}

// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateProductRequest struct {
	CategoryID  int64   `json:"category_id"  binding:"required"`
	Name        string  `json:"name"         binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"        binding:"required,gt=0"`
	Stock       int     `json:"stock"        binding:"gte=0"`
	ImageURL    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"       binding:"required,gt=0"`
	Stock       int     `json:"stock"       binding:"gte=0"`
	ImageURL    string  `json:"image_url"`
}

type ProductFilter struct {
	Search     string   `form:"search"`
	CategoryID *int64   `form:"category_id"`
	MinPrice   *float64 `form:"min_price"`
	MaxPrice   *float64 `form:"max_price"`
	Page       int      `form:"page,default=1"`
	Limit      int      `form:"limit,default=20"`
}
