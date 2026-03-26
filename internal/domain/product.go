package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	CategoryID  uuid.UUID `db:"category_id"`
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
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	CreatedAt time.Time `db:"created_at"`
}

// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateProductRequest struct {
	CategoryID  uuid.UUID `json:"category_id"  binding:"required"`
	Name        string    `json:"name"         binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price"        binding:"required,gt=0"`
	Stock       int       `json:"stock"        binding:"gte=0"`
	ImageURL    string    `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"       binding:"required,gt=0"`
	Stock       int     `json:"stock"       binding:"gte=0"`
	ImageURL    string  `json:"image_url"`
}

type ProductFilter struct {
	Search     string     `form:"search"`
	CategoryID *uuid.UUID `form:"category_id"`
	MinPrice   *float64   `form:"min_price"`
	MaxPrice   *float64   `form:"max_price"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=20"`
}
