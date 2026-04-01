package pb

// ── Order ────────────────────────────────────────────────────────────────────

type OrderItemMessage struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderMessage struct {
	ID         int64              `json:"id"`
	UserID     int64              `json:"user_id"`
	AddressID  int64              `json:"address_id,omitempty"`
	Status     string             `json:"status"`
	TotalPrice float64            `json:"total_price"`
	CreatedAt  int64              `json:"created_at"`
	UpdatedAt  int64              `json:"updated_at"`
	Items      []*OrderItemMessage `json:"items,omitempty"`
}

type CreateOrderItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type CreateOrderRequest struct {
	AddressID int64                    `json:"address_id,omitempty"`
	Items     []*CreateOrderItemRequest `json:"items"`
}

type GetOrderRequest struct {
	ID int64 `json:"id"`
}

type ListOrdersRequest struct {
	Page  int32 `json:"page,omitempty"`
	Limit int32 `json:"limit,omitempty"`
}

type ListOrdersResponse struct {
	Orders []*OrderMessage `json:"orders"`
	Total  int32           `json:"total"`
	Page   int32           `json:"page"`
	Limit  int32           `json:"limit"`
}

type UpdateOrderStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateOrderStatusResponse struct{}

type CancelOrderRequest struct {
	ID int64 `json:"id"`
}

type CancelOrderResponse struct{}
