package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/cache"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

type Handler struct {
	services         *service.Services
	authMgr          auth.Manager
	logger           *slog.Logger
	db               *sqlx.DB
	idempotencyStore *cache.IdempotencyStore
}

func NewHandler(services *service.Services, authMgr auth.Manager, logger *slog.Logger, db *sqlx.DB, idempotencyStore *cache.IdempotencyStore) *Handler {
	return &Handler{services: services, authMgr: authMgr, logger: logger, db: db, idempotencyStore: idempotencyStore}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(h.requestLogger())

	r.GET("/health", h.Health)

	api := r.Group("/api/v1")

	// ── Public: auth ──────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}

	// ── Public: product browsing ──────────────────────────────────────────────
	api.GET("/products", h.ListProducts)
	api.GET("/products/:id", h.GetProductByID)
	api.GET("/categories", h.ListCategories)

	// ── Protected: JWT required ───────────────────────────────────────────────
	protected := api.Group("", middleware.Auth(h.authMgr))
	{
		// Users — list/delete restricted to admin
		users := protected.Group("/users")
		{
			users.GET("", middleware.RequireRole(models.RoleAdmin), h.ListUsers)
			users.GET("/:id", h.GetUserByID)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), h.DeleteUser)
		}

		// Products — write operations for admin and seller
		products := protected.Group("/products")
		{
			products.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleSeller), h.CreateProduct)
			products.PUT("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSeller), h.UpdateProduct)
			products.DELETE("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSeller), h.DeleteProduct)
		}

		// Cart — authenticated customers
		cart := protected.Group("/cart")
		{
			cart.GET("", h.GetCart)
			cart.POST("/items", h.AddToCart)
			cart.PUT("/items/:productId", h.UpdateCartItem)
			cart.DELETE("/items/:productId", h.RemoveFromCart)
			cart.DELETE("", h.ClearCart)
			cart.POST("/checkout", middleware.Idempotency(h.idempotencyStore), h.Checkout)
		}

		// Categories — write operations for admin only
		categories := protected.Group("/categories")
		{
			categories.POST("", middleware.RequireRole(models.RoleAdmin), h.CreateCategory)
			categories.PUT("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleManager), h.UpdateCategory)
			categories.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), h.DeleteCategory)
		}

		// Seller dashboard
		protected.GET("/seller/products", middleware.RequireRole(models.RoleSeller, models.RoleAdmin), h.ListSellerProducts)

		// Orders
		orders := protected.Group("/orders")
		{
			orders.POST("", middleware.Idempotency(h.idempotencyStore), h.CreateOrder)
			orders.GET("", h.ListOrders)
			orders.GET("/:id", h.GetOrderByID)
			orders.PATCH("/:id/cancel", h.CancelOrder)
			orders.PATCH("/:id/status", middleware.RequireRole(models.RoleAdmin, models.RoleManager), h.UpdateOrderStatus)
		}

		// Addresses
		addresses := protected.Group("/addresses")
		{
			addresses.GET("", h.ListAddresses)
			addresses.POST("", h.CreateAddress)
			addresses.PUT("/:id", h.UpdateAddress)
			addresses.DELETE("/:id", h.DeleteAddress)
			addresses.PATCH("/:id/default", h.SetDefaultAddress)
		}

		// Reviews — write operations protected, read is public (registered below)
		productReviews := protected.Group("/products/:id/reviews")
		{
			productReviews.POST("", h.CreateReview)
			productReviews.PUT("/:reviewId", h.UpdateReview)
			productReviews.DELETE("/:reviewId", h.DeleteReview)
		}
	}

	// Public: product reviews
	api.GET("/products/:id/reviews", h.ListReviews)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "route not found"})
	})

	return r
}

func (h *Handler) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		h.logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
		)
	}
}
