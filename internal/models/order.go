package models

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipping  OrderStatus = "shipping"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         int64       `db:"id"          json:"id"`
	UserID     int64       `db:"user_id"      json:"user_id"`
	AddressID  *int64      `db:"address_id"   json:"address_id,omitempty"`
	Status     OrderStatus `db:"status"       json:"status"`
	TotalPrice float64     `db:"total_price"  json:"total_price"`
	CreatedAt  time.Time   `db:"created_at"   json:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"   json:"updated_at"`
	Items      []OrderItem `db:"-"            json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64   `db:"id"`
	OrderID   int64   `db:"order_id"`
	ProductID int64   `db:"product_id"`
	Quantity  int     `db:"quantity"`
	UnitPrice float64 `db:"unit_price"`
}

type CreateOrder struct {
	AddressID *int64            `json:"address_id"`
	Items     []CreateOrderItem `json:"items" binding:"required,min=1,dive"`
}

type CreateOrderItem struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity"   binding:"required,min=1"`
}

type UpdateOrderStatus struct {
	Status OrderStatus `json:"status" binding:"required,oneof=confirmed shipping delivered"`
}
