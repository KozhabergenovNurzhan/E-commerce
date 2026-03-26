package testutil

import (
	"context"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
)

// ── UserRepository mock ───────────────────────────────────────────────────────

type MockUserRepo struct {
	CreateFn      func(ctx context.Context, user *domain.User) error
	FindByIDFn    func(ctx context.Context, id int64) (*domain.User, error)
	FindByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	UpdateFn      func(ctx context.Context, user *domain.User) error
	DeleteFn      func(ctx context.Context, id int64) error
	ListFn        func(ctx context.Context, limit, offset int) ([]*domain.User, int, error)
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFn(ctx, user)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.FindByEmailFn(ctx, email)
}
func (m *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.UpdateFn(ctx, user)
}
func (m *MockUserRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	return m.ListFn(ctx, limit, offset)
}

// ── ProductRepository mock ────────────────────────────────────────────────────

type MockProductRepo struct {
	CreateFn        func(ctx context.Context, p *domain.Product) error
	FindByIDFn      func(ctx context.Context, id int64) (*domain.Product, error)
	UpdateFn        func(ctx context.Context, p *domain.Product) error
	DeleteFn        func(ctx context.Context, id int64) error
	ListFn          func(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error)
	ListBySellerFn  func(ctx context.Context, sellerID int64, f *domain.ProductFilter) ([]*domain.Product, int, error)
	ListCategoriesFn func(ctx context.Context) ([]*domain.Category, error)
}

func (m *MockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	return m.CreateFn(ctx, p)
}
func (m *MockProductRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *MockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	return m.UpdateFn(ctx, p)
}
func (m *MockProductRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockProductRepo) List(ctx context.Context, f *domain.ProductFilter) ([]*domain.Product, int, error) {
	return m.ListFn(ctx, f)
}
func (m *MockProductRepo) ListBySeller(ctx context.Context, sellerID int64, f *domain.ProductFilter) ([]*domain.Product, int, error) {
	return m.ListBySellerFn(ctx, sellerID, f)
}
func (m *MockProductRepo) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return m.ListCategoriesFn(ctx)
}

// ── OrderRepository mock ──────────────────────────────────────────────────────

type MockOrderRepo struct {
	CreateFn       func(ctx context.Context, order *domain.Order) error
	FindByIDFn     func(ctx context.Context, id int64) (*domain.Order, error)
	ListByUserFn   func(ctx context.Context, userID int64, limit, offset int) ([]*domain.Order, int, error)
	UpdateStatusFn func(ctx context.Context, id int64, status domain.OrderStatus) error
}

func (m *MockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	return m.CreateFn(ctx, order)
}
func (m *MockOrderRepo) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *MockOrderRepo) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Order, int, error) {
	return m.ListByUserFn(ctx, userID, limit, offset)
}
func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

// ── TokenRepository mock ──────────────────────────────────────────────────────

type MockTokenRepo struct {
	SaveFn       func(ctx context.Context, t *domain.RefreshToken) error
	FindByHashFn func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeFn     func(ctx context.Context, hash string) error
}

func (m *MockTokenRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	return m.SaveFn(ctx, t)
}
func (m *MockTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	return m.FindByHashFn(ctx, hash)
}
func (m *MockTokenRepo) Revoke(ctx context.Context, hash string) error {
	return m.RevokeFn(ctx, hash)
}

// ── AuthManager mock ──────────────────────────────────────────────────────────

type MockAuthManager struct {
	GenerateAccessTokenFn func(userID int64, role domain.Role) (string, error)
	ValidateAccessTokenFn func(token string) (*auth.Claims, error)
	AccessTTLFn           func() time.Duration
}

func (m *MockAuthManager) GenerateAccessToken(userID int64, role domain.Role) (string, error) {
	return m.GenerateAccessTokenFn(userID, role)
}
func (m *MockAuthManager) ValidateAccessToken(token string) (*auth.Claims, error) {
	return m.ValidateAccessTokenFn(token)
}
func (m *MockAuthManager) AccessTTL() time.Duration {
	return m.AccessTTLFn()
}
