package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

type Handler struct {
	services *service.Services
	authMgr  auth.Manager
	logger   *slog.Logger
}

func NewHandler(services *service.Services, authMgr auth.Manager, logger *slog.Logger) *Handler {
	return &Handler{services: services, authMgr: authMgr, logger: logger}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(h.requestLogger())

	r.GET("/health", h.Health)

	api := r.Group("/api/v1")

	// ── Public: auth ──────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login",    h.Login)
		auth.POST("/refresh",  h.Refresh)
		auth.POST("/logout",   h.Logout)
	}

	// ── Public: product browsing ──────────────────────────────────────────────
	api.GET("/products",     h.ListProducts)
	api.GET("/products/:id", h.GetProductByID)
	api.GET("/categories",   h.ListCategories)

	// ── Protected: JWT required ───────────────────────────────────────────────
	protected := api.Group("", middleware.Auth(h.authMgr))
	{
		// Users — list/delete restricted to admin
		users := protected.Group("/users")
		{
			users.GET("",        middleware.RequireRole(domain.RoleAdmin), h.ListUsers)
			users.GET("/:id",    h.GetUserByID)
			users.PUT("/:id",    h.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole(domain.RoleAdmin), h.DeleteUser)
		}

		// Products — write operations restricted to admin
		products := protected.Group("/products")
		{
			products.POST("",       middleware.RequireRole(domain.RoleAdmin), h.CreateProduct)
			products.PUT("/:id",    middleware.RequireRole(domain.RoleAdmin), h.UpdateProduct)
			products.DELETE("/:id", middleware.RequireRole(domain.RoleAdmin), h.DeleteProduct)
		}

		// Orders — any authenticated user
		orders := protected.Group("/orders")
		{
			orders.POST("",             h.CreateOrder)
			orders.GET("",              h.ListOrders)
			orders.GET("/:id",          h.GetOrderByID)
			orders.PATCH("/:id/cancel", h.CancelOrder)
		}
	}

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
			slog.String("method",    c.Request.Method),
			slog.String("path",      c.Request.URL.Path),
			slog.Int("status",       c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip",        c.ClientIP()),
		)
	}
}
