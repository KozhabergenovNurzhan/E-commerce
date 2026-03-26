package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

// GET /api/v1/products
func (h *Handler) ListProducts(c *gin.Context) {
	var f domain.ProductFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, total, err := h.services.Product.List(c.Request.Context(), &f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, products, &response.Meta{Page: f.Page, Limit: f.Limit, Total: total})
}

// GET /api/v1/products/:id
func (h *Handler) GetProductByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	p, err := h.services.Product.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

// POST /api/v1/products
func (h *Handler) CreateProduct(c *gin.Context) {
	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.services.Product.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("create product failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}
	response.Created(c, p)
}

// PUT /api/v1/products/:id
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req domain.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.services.Product.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

// DELETE /api/v1/products/:id
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.services.Product.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// GET /api/v1/categories
func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.services.Product.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cats)
}
