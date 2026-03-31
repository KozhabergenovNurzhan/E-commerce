package service

import (
	"context"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type UserServiceI interface {
	Register(ctx context.Context, req *models.Register) (*models.User, error)
	Login(ctx context.Context, req *models.Login) (*models.UserRecord, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, id int64, firstName, lastName string) (*models.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, limit int) ([]*models.User, int, error)
}

type ProductServiceI interface {
	Create(ctx context.Context, sellerID *int64, req *models.CreateProduct) (*models.Product, error)
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	Update(ctx context.Context, id int64, req *models.UpdateProduct) (*models.Product, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.ProductFilter) ([]*models.Product, int, error)
	ListBySeller(ctx context.Context, sellerID int64, filter *models.ProductFilter) ([]*models.Product, int, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
	CreateCategory(ctx context.Context, req *models.CreateCategory) (*models.Category, error)
	UpdateCategory(ctx context.Context, id int64, req *models.UpdateCategory) (*models.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

type OrderServiceI interface {
	Create(ctx context.Context, userID int64, req *models.CreateOrder) (*models.Order, error)
	GetByID(ctx context.Context, id int64) (*models.Order, error)
	ListByUser(ctx context.Context, userID int64, page, limit int) ([]*models.Order, int, error)
	UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error
	Cancel(ctx context.Context, id int64) error
}

type TokenServiceI interface {
	GenerateTokenPair(ctx context.Context, userID int64, role models.Role) (*models.AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (*models.AuthTokens, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type CartServiceI interface {
	AddItem(ctx context.Context, userID int64, req *models.AddToCart) error
	GetCart(ctx context.Context, userID int64) (*models.Cart, error)
	UpdateItem(ctx context.Context, userID, productID int64, req *models.UpdateCartItem) error
	RemoveItem(ctx context.Context, userID, productID int64) error
	Clear(ctx context.Context, userID int64) error
}

type ReviewServiceI interface {
	Create(ctx context.Context, userID, productID int64, req *models.CreateReview) (*models.Review, error)
	ListByProduct(ctx context.Context, productID int64, page, limit int) ([]*models.Review, int, *models.ProductRating, error)
	Update(ctx context.Context, reviewID, userID int64, req *models.UpdateReview) (*models.Review, error)
	Delete(ctx context.Context, reviewID, callerID int64, callerRole models.Role) error
}

type AddressServiceI interface {
	Create(ctx context.Context, userID int64, req *models.CreateAddress) (*models.Address, error)
	List(ctx context.Context, userID int64) ([]*models.Address, error)
	Update(ctx context.Context, addressID, userID int64, req *models.UpdateAddress) (*models.Address, error)
	Delete(ctx context.Context, addressID, userID int64) error
	SetDefault(ctx context.Context, addressID, userID int64) (*models.Address, error)
}
