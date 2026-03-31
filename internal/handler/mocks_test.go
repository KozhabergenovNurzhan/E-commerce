package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

// ── Mock service types ────────────────────────────────────────────────────────

type MockUserService struct {
	RegisterFn func(ctx context.Context, req *models.Register) (*models.User, error)
	LoginFn    func(ctx context.Context, req *models.Login) (*models.UserRecord, error)
	GetByIDFn  func(ctx context.Context, id int64) (*models.User, error)
	UpdateFn   func(ctx context.Context, id int64, firstName, lastName string) (*models.User, error)
	DeleteFn   func(ctx context.Context, id int64) error
	ListFn     func(ctx context.Context, page, limit int) ([]*models.User, int, error)
}

var _ service.UserServiceI = (*MockUserService)(nil)

func (m *MockUserService) Register(ctx context.Context, req *models.Register) (*models.User, error) {
	return m.RegisterFn(ctx, req)
}
func (m *MockUserService) Login(ctx context.Context, req *models.Login) (*models.UserRecord, error) {
	return m.LoginFn(ctx, req)
}
func (m *MockUserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockUserService) Update(ctx context.Context, id int64, firstName, lastName string) (*models.User, error) {
	return m.UpdateFn(ctx, id, firstName, lastName)
}
func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockUserService) List(ctx context.Context, page, limit int) ([]*models.User, int, error) {
	return m.ListFn(ctx, page, limit)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockProductService struct {
	CreateFn         func(ctx context.Context, sellerID *int64, req *models.CreateProduct) (*models.Product, error)
	GetByIDFn        func(ctx context.Context, id int64) (*models.Product, error)
	UpdateFn         func(ctx context.Context, id int64, req *models.UpdateProduct) (*models.Product, error)
	DeleteFn         func(ctx context.Context, id int64) error
	ListFn           func(ctx context.Context, filter *models.ProductFilter) ([]*models.Product, int, error)
	ListBySellerFn   func(ctx context.Context, sellerID int64, filter *models.ProductFilter) ([]*models.Product, int, error)
	ListCategoriesFn func(ctx context.Context) ([]*models.Category, error)
	CreateCategoryFn func(ctx context.Context, req *models.CreateCategory) (*models.Category, error)
	UpdateCategoryFn func(ctx context.Context, id int64, req *models.UpdateCategory) (*models.Category, error)
	DeleteCategoryFn func(ctx context.Context, id int64) error
}

var _ service.ProductServiceI = (*MockProductService)(nil)

func (m *MockProductService) Create(ctx context.Context, sellerID *int64, req *models.CreateProduct) (*models.Product, error) {
	return m.CreateFn(ctx, sellerID, req)
}
func (m *MockProductService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockProductService) Update(ctx context.Context, id int64, req *models.UpdateProduct) (*models.Product, error) {
	return m.UpdateFn(ctx, id, req)
}
func (m *MockProductService) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockProductService) List(ctx context.Context, filter *models.ProductFilter) ([]*models.Product, int, error) {
	return m.ListFn(ctx, filter)
}
func (m *MockProductService) ListBySeller(ctx context.Context, sellerID int64, filter *models.ProductFilter) ([]*models.Product, int, error) {
	return m.ListBySellerFn(ctx, sellerID, filter)
}
func (m *MockProductService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return m.ListCategoriesFn(ctx)
}
func (m *MockProductService) CreateCategory(ctx context.Context, req *models.CreateCategory) (*models.Category, error) {
	return m.CreateCategoryFn(ctx, req)
}
func (m *MockProductService) UpdateCategory(ctx context.Context, id int64, req *models.UpdateCategory) (*models.Category, error) {
	return m.UpdateCategoryFn(ctx, id, req)
}
func (m *MockProductService) DeleteCategory(ctx context.Context, id int64) error {
	return m.DeleteCategoryFn(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockOrderService struct {
	CreateFn       func(ctx context.Context, userID int64, req *models.CreateOrder) (*models.Order, error)
	GetByIDFn      func(ctx context.Context, id int64) (*models.Order, error)
	ListByUserFn   func(ctx context.Context, userID int64, page, limit int) ([]*models.Order, int, error)
	UpdateStatusFn func(ctx context.Context, id int64, status models.OrderStatus) error
	CancelFn       func(ctx context.Context, id int64) error
}

var _ service.OrderServiceI = (*MockOrderService)(nil)

func (m *MockOrderService) Create(ctx context.Context, userID int64, req *models.CreateOrder) (*models.Order, error) {
	return m.CreateFn(ctx, userID, req)
}
func (m *MockOrderService) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockOrderService) ListByUser(ctx context.Context, userID int64, page, limit int) ([]*models.Order, int, error) {
	return m.ListByUserFn(ctx, userID, page, limit)
}
func (m *MockOrderService) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}
func (m *MockOrderService) Cancel(ctx context.Context, id int64) error {
	return m.CancelFn(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockTokenService struct {
	GenerateTokenPairFn func(ctx context.Context, userID int64, role models.Role) (*models.AuthTokens, error)
	RefreshFn           func(ctx context.Context, refreshToken string) (*models.AuthTokens, error)
	RevokeFn            func(ctx context.Context, refreshToken string) error
}

var _ service.TokenServiceI = (*MockTokenService)(nil)

func (m *MockTokenService) GenerateTokenPair(ctx context.Context, userID int64, role models.Role) (*models.AuthTokens, error) {
	return m.GenerateTokenPairFn(ctx, userID, role)
}
func (m *MockTokenService) Refresh(ctx context.Context, refreshToken string) (*models.AuthTokens, error) {
	return m.RefreshFn(ctx, refreshToken)
}
func (m *MockTokenService) Revoke(ctx context.Context, refreshToken string) error {
	return m.RevokeFn(ctx, refreshToken)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockCartService struct {
	AddItemFn    func(ctx context.Context, userID int64, req *models.AddToCart) error
	GetCartFn    func(ctx context.Context, userID int64) (*models.Cart, error)
	UpdateItemFn func(ctx context.Context, userID, productID int64, req *models.UpdateCartItem) error
	RemoveItemFn func(ctx context.Context, userID, productID int64) error
	ClearFn      func(ctx context.Context, userID int64) error
}

var _ service.CartServiceI = (*MockCartService)(nil)

func (m *MockCartService) AddItem(ctx context.Context, userID int64, req *models.AddToCart) error {
	return m.AddItemFn(ctx, userID, req)
}
func (m *MockCartService) GetCart(ctx context.Context, userID int64) (*models.Cart, error) {
	return m.GetCartFn(ctx, userID)
}
func (m *MockCartService) UpdateItem(ctx context.Context, userID, productID int64, req *models.UpdateCartItem) error {
	return m.UpdateItemFn(ctx, userID, productID, req)
}
func (m *MockCartService) RemoveItem(ctx context.Context, userID, productID int64) error {
	return m.RemoveItemFn(ctx, userID, productID)
}
func (m *MockCartService) Clear(ctx context.Context, userID int64) error {
	return m.ClearFn(ctx, userID)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockReviewService struct {
	CreateFn        func(ctx context.Context, userID, productID int64, req *models.CreateReview) (*models.Review, error)
	ListByProductFn func(ctx context.Context, productID int64, page, limit int) ([]*models.Review, int, *models.ProductRating, error)
	UpdateFn        func(ctx context.Context, reviewID, userID int64, req *models.UpdateReview) (*models.Review, error)
	DeleteFn        func(ctx context.Context, reviewID, callerID int64, callerRole models.Role) error
}

var _ service.ReviewServiceI = (*MockReviewService)(nil)

func (m *MockReviewService) Create(ctx context.Context, userID, productID int64, req *models.CreateReview) (*models.Review, error) {
	return m.CreateFn(ctx, userID, productID, req)
}
func (m *MockReviewService) ListByProduct(ctx context.Context, productID int64, page, limit int) ([]*models.Review, int, *models.ProductRating, error) {
	return m.ListByProductFn(ctx, productID, page, limit)
}
func (m *MockReviewService) Update(ctx context.Context, reviewID, userID int64, req *models.UpdateReview) (*models.Review, error) {
	return m.UpdateFn(ctx, reviewID, userID, req)
}
func (m *MockReviewService) Delete(ctx context.Context, reviewID, callerID int64, callerRole models.Role) error {
	return m.DeleteFn(ctx, reviewID, callerID, callerRole)
}

// ─────────────────────────────────────────────────────────────────────────────

type MockAddressService struct {
	CreateFn     func(ctx context.Context, userID int64, req *models.CreateAddress) (*models.Address, error)
	ListFn       func(ctx context.Context, userID int64) ([]*models.Address, error)
	UpdateFn     func(ctx context.Context, addressID, userID int64, req *models.UpdateAddress) (*models.Address, error)
	DeleteFn     func(ctx context.Context, addressID, userID int64) error
	SetDefaultFn func(ctx context.Context, addressID, userID int64) (*models.Address, error)
}

var _ service.AddressServiceI = (*MockAddressService)(nil)

func (m *MockAddressService) Create(ctx context.Context, userID int64, req *models.CreateAddress) (*models.Address, error) {
	return m.CreateFn(ctx, userID, req)
}
func (m *MockAddressService) List(ctx context.Context, userID int64) ([]*models.Address, error) {
	return m.ListFn(ctx, userID)
}
func (m *MockAddressService) Update(ctx context.Context, addressID, userID int64, req *models.UpdateAddress) (*models.Address, error) {
	return m.UpdateFn(ctx, addressID, userID, req)
}
func (m *MockAddressService) Delete(ctx context.Context, addressID, userID int64) error {
	return m.DeleteFn(ctx, addressID, userID)
}
func (m *MockAddressService) SetDefault(ctx context.Context, addressID, userID int64) (*models.Address, error) {
	return m.SetDefaultFn(ctx, addressID, userID)
}

// ── Test helpers ──────────────────────────────────────────────────────────────

func newTestContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

func setAuth(c *gin.Context, userID int64, role models.Role) {
	c.Set(middleware.CtxUserID, userID)
	c.Set(middleware.CtxUserRole, role)
}

func decodeBodyMap(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func newTestHandler(svcs *service.Services) *Handler {
	return &Handler{
		services: svcs,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}
