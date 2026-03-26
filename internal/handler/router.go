package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

type Router struct {
	auth     *AuthHandler
	user     *UserHandler
	product  *ProductHandler
	order    *OrderHandler
	tokenSvc service.TokenService
	logger   *slog.Logger
}

func NewRouter(
	auth *AuthHandler,
	user *UserHandler,
	product *ProductHandler,
	order *OrderHandler,
	tokenSvc service.TokenService,
	logger *slog.Logger,
) *Router {
	return &Router{
		auth:     auth,
		user:     user,
		product:  product,
		order:    order,
		tokenSvc: tokenSvc,
		logger:   logger,
	}
}

func (r *Router) Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(r.requestLogger())

	engine.GET("/health", r.user.Health)

	v1 := engine.Group("/api/v1")

	// ── Public: auth ──────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/register", r.auth.Register)
		auth.POST("/login",    r.auth.Login)
		auth.POST("/refresh",  r.auth.Refresh)
		auth.POST("/logout",   r.auth.Logout)
	}

	// ── Public: product browsing ──────────────────────────────────────────────
	v1.GET("/products",     r.product.List)
	v1.GET("/products/:id", r.product.GetByID)
	v1.GET("/categories",   r.product.ListCategories)

	// ── Protected: JWT required ───────────────────────────────────────────────
	protected := v1.Group("", JWTMiddleware(r.tokenSvc))
	{
		users := protected.Group("/users")
		{
			users.GET("",          r.user.List)
			users.GET("/:id",      r.user.GetByID)
			users.PUT("/:id",      r.user.Update)
			users.DELETE("/:id",   r.user.Delete)
		}

		products := protected.Group("/products")
		{
			products.POST("",       r.product.Create)
			products.PUT("/:id",    r.product.Update)
			products.DELETE("/:id", r.product.Delete)
		}

		orders := protected.Group("/orders")
		{
			orders.POST("",              r.order.Create)
			orders.GET("",               r.order.List)
			orders.GET("/:id",           r.order.GetByID)
			orders.PATCH("/:id/cancel",  r.order.Cancel)
		}
	}

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "route not found"})
	})

	return engine
}

func (r *Router) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		r.logger.Info("request",
			slog.String("method",  c.Request.Method),
			slog.String("path",    c.Request.URL.Path),
			slog.Int("status",     c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip",      c.ClientIP()),
		)
	}
}
