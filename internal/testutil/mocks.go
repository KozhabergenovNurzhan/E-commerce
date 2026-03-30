package testutil

import (
	"context"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

// ── UserRepository mock ───────────────────────────────────────────────────────

type MockUserRepo struct {
	CreateFn      func(ctx context.Context, user *models.UserRecord) error
	FindByIDFn    func(ctx context.Context, id int64) (*models.UserRecord, error)
	FindByEmailFn func(ctx context.Context, email string) (*models.UserRecord, error)
	UpdateFn      func(ctx context.Context, user *models.UserRecord) error
	DeleteFn      func(ctx context.Context, id int64) error
	ListFn        func(ctx context.Context, limit, offset int) ([]*models.UserRecord, int, error)
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.UserRecord) error {
	return m.CreateFn(ctx, user)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id int64) (*models.UserRecord, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.UserRecord, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.UserRecord) error {
	return m.UpdateFn(ctx, user)
}

func (m *MockUserRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*models.UserRecord, int, error) {
	return m.ListFn(ctx, limit, offset)
}

// ── ProductRepository mock ────────────────────────────────────────────────────

type MockProductRepo struct {
	CreateFn         func(ctx context.Context, p *models.Product) error
	FindByIDFn       func(ctx context.Context, id int64) (*models.Product, error)
	UpdateFn         func(ctx context.Context, p *models.Product) error
	DeleteFn         func(ctx context.Context, id int64) error
	ListFn           func(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error)
	ListBySellerFn   func(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error)
	ListCategoriesFn func(ctx context.Context) ([]*models.Category, error)
	CreateCategoryFn func(ctx context.Context, c *models.Category) error
	UpdateCategoryFn func(ctx context.Context, c *models.Category) error
	DeleteCategoryFn func(ctx context.Context, id int64) error
}

func (m *MockProductRepo) Create(ctx context.Context, p *models.Product) error {
	return m.CreateFn(ctx, p)
}

func (m *MockProductRepo) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockProductRepo) Update(ctx context.Context, p *models.Product) error {
	return m.UpdateFn(ctx, p)
}

func (m *MockProductRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockProductRepo) List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
	return m.ListFn(ctx, f)
}

func (m *MockProductRepo) ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error) {
	return m.ListBySellerFn(ctx, sellerID, f)
}

func (m *MockProductRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return m.ListCategoriesFn(ctx)
}

func (m *MockProductRepo) CreateCategory(ctx context.Context, c *models.Category) error {
	return m.CreateCategoryFn(ctx, c)
}

func (m *MockProductRepo) UpdateCategory(ctx context.Context, c *models.Category) error {
	return m.UpdateCategoryFn(ctx, c)
}

func (m *MockProductRepo) DeleteCategory(ctx context.Context, id int64) error {
	return m.DeleteCategoryFn(ctx, id)
}

// ── OrderRepository mock ──────────────────────────────────────────────────────

type MockOrderRepo struct {
	CreateFn       func(ctx context.Context, order *models.Order) error
	FindByIDFn     func(ctx context.Context, id int64) (*models.Order, error)
	ListByUserFn   func(ctx context.Context, userID int64, limit, offset int) ([]*models.Order, int, error)
	UpdateStatusFn func(ctx context.Context, id int64, status models.OrderStatus) error
}

func (m *MockOrderRepo) Create(ctx context.Context, order *models.Order) error {
	return m.CreateFn(ctx, order)
}

func (m *MockOrderRepo) FindByID(ctx context.Context, id int64) (*models.Order, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockOrderRepo) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Order, int, error) {
	return m.ListByUserFn(ctx, userID, limit, offset)
}

func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

// ── CartRepository mock ───────────────────────────────────────────────────────

type MockCartRepo struct {
	UpsertFn       func(ctx context.Context, item *models.CartItemRecord) error
	FindByUserIDFn func(ctx context.Context, userID int64) ([]*models.CartItemRecord, error)
	DeleteFn       func(ctx context.Context, userID, productID int64) error
	ClearFn        func(ctx context.Context, userID int64) error
}

func (m *MockCartRepo) Upsert(ctx context.Context, item *models.CartItemRecord) error {
	return m.UpsertFn(ctx, item)
}

func (m *MockCartRepo) FindByUserID(ctx context.Context, userID int64) ([]*models.CartItemRecord, error) {
	return m.FindByUserIDFn(ctx, userID)
}

func (m *MockCartRepo) Delete(ctx context.Context, userID, productID int64) error {
	return m.DeleteFn(ctx, userID, productID)
}

func (m *MockCartRepo) Clear(ctx context.Context, userID int64) error {
	return m.ClearFn(ctx, userID)
}

// ── TokenRepository mock ──────────────────────────────────────────────────────

type MockTokenRepo struct {
	SaveFn       func(ctx context.Context, t *models.RefreshToken) error
	FindByHashFn func(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeFn     func(ctx context.Context, hash string) error
}

func (m *MockTokenRepo) Save(ctx context.Context, t *models.RefreshToken) error {
	return m.SaveFn(ctx, t)
}

func (m *MockTokenRepo) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return m.FindByHashFn(ctx, hash)
}

func (m *MockTokenRepo) Revoke(ctx context.Context, hash string) error {
	return m.RevokeFn(ctx, hash)
}

// ── ReviewRepository mock ─────────────────────────────────────────────────────

type MockReviewRepo struct {
	CreateFn        func(ctx context.Context, r *models.Review) error
	FindByIDFn      func(ctx context.Context, id int64) (*models.Review, error)
	ListByProductFn func(ctx context.Context, productID int64, limit, offset int) ([]*models.Review, int, error)
	UpdateFn        func(ctx context.Context, r *models.Review) error
	DeleteFn        func(ctx context.Context, id int64) error
	GetRatingFn     func(ctx context.Context, productID int64) (*models.ProductRating, error)
}

func (m *MockReviewRepo) Create(ctx context.Context, r *models.Review) error {
	return m.CreateFn(ctx, r)
}

func (m *MockReviewRepo) FindByID(ctx context.Context, id int64) (*models.Review, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockReviewRepo) ListByProduct(ctx context.Context, productID int64, limit, offset int) ([]*models.Review, int, error) {
	return m.ListByProductFn(ctx, productID, limit, offset)
}

func (m *MockReviewRepo) Update(ctx context.Context, r *models.Review) error {
	return m.UpdateFn(ctx, r)
}

func (m *MockReviewRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockReviewRepo) GetRating(ctx context.Context, productID int64) (*models.ProductRating, error) {
	return m.GetRatingFn(ctx, productID)
}

// ── AddressRepository mock ────────────────────────────────────────────────────

type MockAddressRepo struct {
	CreateFn       func(ctx context.Context, a *models.Address) error
	FindByIDFn     func(ctx context.Context, id int64) (*models.Address, error)
	ListByUserFn   func(ctx context.Context, userID int64) ([]*models.Address, error)
	UpdateFn       func(ctx context.Context, a *models.Address) error
	DeleteFn       func(ctx context.Context, id int64) error
	ClearDefaultFn func(ctx context.Context, userID int64) error
}

func (m *MockAddressRepo) Create(ctx context.Context, a *models.Address) error {
	return m.CreateFn(ctx, a)
}

func (m *MockAddressRepo) FindByID(ctx context.Context, id int64) (*models.Address, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockAddressRepo) ListByUser(ctx context.Context, userID int64) ([]*models.Address, error) {
	return m.ListByUserFn(ctx, userID)
}

func (m *MockAddressRepo) Update(ctx context.Context, a *models.Address) error {
	return m.UpdateFn(ctx, a)
}

func (m *MockAddressRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockAddressRepo) ClearDefault(ctx context.Context, userID int64) error {
	return m.ClearDefaultFn(ctx, userID)
}

// ── AuthManager mock ──────────────────────────────────────────────────────────

type MockAuthManager struct {
	GenerateAccessTokenFn func(userID int64, role models.Role) (string, error)
	ValidateAccessTokenFn func(token string) (*auth.Claims, error)
	AccessTTLFn           func() time.Duration
}

func (m *MockAuthManager) GenerateAccessToken(userID int64, role models.Role) (string, error) {
	return m.GenerateAccessTokenFn(userID, role)
}

func (m *MockAuthManager) ValidateAccessToken(token string) (*auth.Claims, error) {
	return m.ValidateAccessTokenFn(token)
}

func (m *MockAuthManager) AccessTTL() time.Duration {
	return m.AccessTTLFn()
}
