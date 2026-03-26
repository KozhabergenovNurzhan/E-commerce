package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

type AuthHandler struct {
	userSvc  service.UserService
	tokenSvc service.TokenService
	logger   *slog.Logger
}

func NewAuthHandler(userSvc service.UserService, tokenSvc service.TokenService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, tokenSvc: tokenSvc, logger: logger}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("register failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}

	tokens, err := h.tokenSvc.GenerateTokenPair(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, domain.AuthResponse{User: user, Tokens: tokens})
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	tokens, err := h.tokenSvc.GenerateTokenPair(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, domain.AuthResponse{User: user.ToResponse(), Tokens: tokens})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, err := h.tokenSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tokens)
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.tokenSvc.Revoke(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
