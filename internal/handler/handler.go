package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

type Handler struct {
	services *service.Services
	logger   *slog.Logger
}

func NewHandler(services *service.Services, logger *slog.Logger) *Handler {
	return &Handler{services: services, logger: logger}
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
	protected := api.Group("", JWTMiddleware(h.services.Token))
	{
		users := protected.Group("/users")
		{
			users.GET("",          h.ListUsers)
			users.GET("/:id",      h.GetUserByID)
			users.PUT("/:id",      h.UpdateUser)
			users.DELETE("/:id",   h.DeleteUser)
		}

		products := protected.Group("/products")
		{
			products.POST("",       h.CreateProduct)
			products.PUT("/:id",    h.UpdateProduct)
			products.DELETE("/:id", h.DeleteProduct)
		}

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
			slog.String("method",   c.Request.Method),
			slog.String("path",     c.Request.URL.Path),
			slog.Int("status",      c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip",       c.ClientIP()),
		)
	}
}
