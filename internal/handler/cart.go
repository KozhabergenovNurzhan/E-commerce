package handler

import (
	"strconv"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

// GET /api/v1/cart
func (h *Handler) GetCart(c *gin.Context) {
	userID := middleware.MustUserID(c)

	cart, err := h.services.Cart.GetCart(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cart)
}

// POST /api/v1/cart/items
func (h *Handler) AddToCart(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req models.AddToCart
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.services.Cart.AddItem(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// PUT /api/v1/cart/items/:productId
func (h *Handler) UpdateCartItem(c *gin.Context) {
	userID := middleware.MustUserID(c)

	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req models.UpdateCartItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.services.Cart.UpdateItem(c.Request.Context(), userID, productID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// DELETE /api/v1/cart/items/:productId
func (h *Handler) RemoveFromCart(c *gin.Context) {
	userID := middleware.MustUserID(c)

	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.services.Cart.RemoveItem(c.Request.Context(), userID, productID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// POST /api/v1/cart/checkout
func (h *Handler) Checkout(c *gin.Context) {
	userID := middleware.MustUserID(c)
	ctx := c.Request.Context()

	cart, err := h.services.Cart.GetCart(ctx, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if len(cart.Items) == 0 {
		response.BadRequest(c, "cart is empty")
		return
	}

	items := make([]models.CreateOrderItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, models.CreateOrderItem{
			ProductID: item.Product.ID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.services.Order.Create(ctx, userID, &models.CreateOrder{Items: items})
	if err != nil {
		response.Error(c, err)
		return
	}

	_ = h.services.Cart.Clear(ctx, userID)

	response.Created(c, order)
}

// DELETE /api/v1/cart
func (h *Handler) ClearCart(c *gin.Context) {
	userID := middleware.MustUserID(c)

	if err := h.services.Cart.Clear(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
