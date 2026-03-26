package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

// POST /api/v1/orders
func (h *Handler) CreateOrder(c *gin.Context) {
	userID := mustUserID(c)

	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.services.Order.Create(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("create order failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}
	response.Created(c, order)
}

// GET /api/v1/orders
func (h *Handler) ListOrders(c *gin.Context) {
	userID := mustUserID(c)

	var q struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=20"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orders, total, err := h.services.Order.ListByUser(c.Request.Context(), userID, q.Page, q.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, orders, &response.Meta{Page: q.Page, Limit: q.Limit, Total: total})
}

// GET /api/v1/orders/:id
func (h *Handler) GetOrderByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	order, err := h.services.Order.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, order)
}

// PATCH /api/v1/orders/:id/cancel
func (h *Handler) CancelOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	if err := h.services.Order.Cancel(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
