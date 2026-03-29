package handler

import (
	"strconv"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// GET /api/v1/addresses
func (h *Handler) ListAddresses(c *gin.Context) {
	userID := middleware.MustUserID(c)

	addresses, err := h.services.Address.List(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, addresses)
}

// POST /api/v1/addresses
func (h *Handler) CreateAddress(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req models.CreateAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	address, err := h.services.Address.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, address)
}

// PUT /api/v1/addresses/:id
func (h *Handler) UpdateAddress(c *gin.Context) {
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	userID := middleware.MustUserID(c)

	var req models.UpdateAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	address, err := h.services.Address.Update(c.Request.Context(), addressID, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, address)
}

// DELETE /api/v1/addresses/:id
func (h *Handler) DeleteAddress(c *gin.Context) {
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	userID := middleware.MustUserID(c)

	if err := h.services.Address.Delete(c.Request.Context(), addressID, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// PATCH /api/v1/addresses/:id/default
func (h *Handler) SetDefaultAddress(c *gin.Context) {
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	userID := middleware.MustUserID(c)

	address, err := h.services.Address.SetDefault(c.Request.Context(), addressID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, address)
}
