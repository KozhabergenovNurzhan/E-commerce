package pb

// ── Product ─────────────────────────────────────────────────────────────────

type ProductMessage struct {
	ID          int64   `json:"id"`
	CategoryID  int64   `json:"category_id"`
	SellerID    int64   `json:"seller_id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
	CreatedAt   int64   `json:"created_at"`
}

type ListProductsRequest struct {
	Search     string  `json:"search,omitempty"`
	CategoryID int64   `json:"category_id,omitempty"`
	MinPrice   float64 `json:"min_price,omitempty"`
	MaxPrice   float64 `json:"max_price,omitempty"`
	Page       int32   `json:"page,omitempty"`
	Limit      int32   `json:"limit,omitempty"`
}

type ListProductsResponse struct {
	Products []*ProductMessage `json:"products"`
	Total    int32             `json:"total"`
	Page     int32             `json:"page"`
	Limit    int32             `json:"limit"`
}

type GetProductRequest struct {
	ID int64 `json:"id"`
}

type CreateProductRequest struct {
	CategoryID  int64   `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	ImageURL    string  `json:"image_url"`
}

type DeleteProductRequest struct {
	ID int64 `json:"id"`
}

type DeleteProductResponse struct{}

type CategoryMessage struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt int64  `json:"created_at"`
}

type ListCategoriesRequest struct{}

type ListCategoriesResponse struct {
	Categories []*CategoryMessage `json:"categories"`
}
