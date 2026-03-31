package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
)

var errDB = errors.New("db error")

func assertAppErr(t *testing.T, err error, expectedCode int, expectedMessage string, expectedWrapped error) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedCode, appErr.Code)
	assert.Equal(t, expectedMessage, appErr.Message)
	if expectedWrapped != nil {
		assert.ErrorIs(t, err, expectedWrapped)
	}
}

// ── UserRepository mock ───────────────────────────────────────────────────────

type MockUserRepo struct{ mock.Mock }

var _ repository.UserRepository = (*MockUserRepo)(nil)

func (m *MockUserRepo) Create(ctx context.Context, user *models.UserRecord) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id int64) (*models.UserRecord, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.UserRecord), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.UserRecord, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*models.UserRecord), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) Update(ctx context.Context, user *models.UserRecord) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*models.UserRecord, int, error) {
	args := m.Called(ctx, limit, offset)
	if v := args.Get(0); v != nil {
		return v.([]*models.UserRecord), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

// ── ProductRepository mock ────────────────────────────────────────────────────

type MockProductRepo struct{ mock.Mock }

var _ repository.ProductRepository = (*MockProductRepo)(nil)

func (m *MockProductRepo) Create(ctx context.Context, p *models.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}
func (m *MockProductRepo) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockProductRepo) Update(ctx context.Context, p *models.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}
func (m *MockProductRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockProductRepo) List(ctx context.Context, f *models.ProductFilter) ([]*models.Product, int, error) {
	args := m.Called(ctx, f)
	if v := args.Get(0); v != nil {
		return v.([]*models.Product), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *MockProductRepo) ListBySeller(ctx context.Context, sellerID int64, f *models.ProductFilter) ([]*models.Product, int, error) {
	args := m.Called(ctx, sellerID, f)
	if v := args.Get(0); v != nil {
		return v.([]*models.Product), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *MockProductRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.([]*models.Category), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockProductRepo) CreateCategory(ctx context.Context, c *models.Category) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockProductRepo) UpdateCategory(ctx context.Context, c *models.Category) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockProductRepo) DeleteCategory(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ── OrderRepository mock ──────────────────────────────────────────────────────

type MockOrderRepo struct{ mock.Mock }

var _ repository.OrderRepository = (*MockOrderRepo)(nil)

func (m *MockOrderRepo) Create(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *MockOrderRepo) FindByID(ctx context.Context, id int64) (*models.Order, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepo) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Order, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if v := args.Get(0); v != nil {
		return v.([]*models.Order), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// ── CartRepository mock ───────────────────────────────────────────────────────

type MockCartRepo struct{ mock.Mock }

var _ repository.CartRepository = (*MockCartRepo)(nil)

func (m *MockCartRepo) Upsert(ctx context.Context, item *models.CartItemRecord) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}
func (m *MockCartRepo) FindByUserID(ctx context.Context, userID int64) ([]*models.CartItemRecord, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]*models.CartItemRecord), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCartRepo) Delete(ctx context.Context, userID, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}
func (m *MockCartRepo) Clear(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ── TokenRepository mock ──────────────────────────────────────────────────────

type MockTokenRepo struct{ mock.Mock }

var _ repository.TokenRepository = (*MockTokenRepo)(nil)

func (m *MockTokenRepo) Save(ctx context.Context, t *models.RefreshToken) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}
func (m *MockTokenRepo) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if v := args.Get(0); v != nil {
		return v.(*models.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTokenRepo) Revoke(ctx context.Context, hash string) error {
	args := m.Called(ctx, hash)
	return args.Error(0)
}

// ── ReviewRepository mock ─────────────────────────────────────────────────────

type MockReviewRepo struct{ mock.Mock }

var _ repository.ReviewRepository = (*MockReviewRepo)(nil)

func (m *MockReviewRepo) Create(ctx context.Context, r *models.Review) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockReviewRepo) FindByID(ctx context.Context, id int64) (*models.Review, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Review), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockReviewRepo) ListByProduct(ctx context.Context, productID int64, limit, offset int) ([]*models.Review, int, error) {
	args := m.Called(ctx, productID, limit, offset)
	if v := args.Get(0); v != nil {
		return v.([]*models.Review), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *MockReviewRepo) Update(ctx context.Context, r *models.Review) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockReviewRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockReviewRepo) GetRating(ctx context.Context, productID int64) (*models.ProductRating, error) {
	args := m.Called(ctx, productID)
	if v := args.Get(0); v != nil {
		return v.(*models.ProductRating), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── AddressRepository mock ────────────────────────────────────────────────────

type MockAddressRepo struct{ mock.Mock }

var _ repository.AddressRepository = (*MockAddressRepo)(nil)

func (m *MockAddressRepo) Create(ctx context.Context, a *models.Address) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}
func (m *MockAddressRepo) FindByID(ctx context.Context, id int64) (*models.Address, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Address), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAddressRepo) ListByUser(ctx context.Context, userID int64) ([]*models.Address, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]*models.Address), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAddressRepo) Update(ctx context.Context, a *models.Address) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}
func (m *MockAddressRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockAddressRepo) ClearDefault(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ── AuthManager mock ──────────────────────────────────────────────────────────

type MockAuthManager struct{ mock.Mock }

var _ auth.Manager = (*MockAuthManager)(nil)

func (m *MockAuthManager) GenerateAccessToken(userID int64, role models.Role) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}
func (m *MockAuthManager) ValidateAccessToken(token string) (*auth.Claims, error) {
	args := m.Called(token)
	if v := args.Get(0); v != nil {
		return v.(*auth.Claims), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAuthManager) AccessTTL() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}
