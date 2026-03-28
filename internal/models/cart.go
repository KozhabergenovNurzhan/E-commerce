package models

import "time"

type CartItemRecord struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	ProductID int64     `db:"product_id"`
	Quantity  int       `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CartItem struct {
	ID       int64    `json:"id"`
	Product  *Product `json:"product"`
	Quantity int      `json:"quantity"`
	Subtotal float64  `json:"subtotal"`
}

type Cart struct {
	Items      []*CartItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
}

type AddToCart struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity"   binding:"required,min=1"`
}

type UpdateCartItem struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}
