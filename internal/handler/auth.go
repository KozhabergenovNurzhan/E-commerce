package handler

import (
	"log/slog"
	"net/http"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

// POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req models.Register
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.services.User.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("register failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}

	tokens, err := h.services.Token.GenerateTokenPair(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, models.Auth{User: user, Tokens: tokens})
}

// POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req models.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.services.User.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	tokens, err := h.services.Token.GenerateTokenPair(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, models.Auth{User: user.ToResponse(), Tokens: tokens})
}

// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, err := h.services.Token.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tokens)
}

// POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.services.Token.Revoke(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// GET /health
func (h *Handler) Health(c *gin.Context) {
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unavailable",
			"service": "ecommerce",
			"db":      "unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "ecommerce",
		"db":      "ok",
	})
}
