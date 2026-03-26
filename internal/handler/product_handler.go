package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

type ProductHandler struct {
	svc    service.ProductService
	logger *slog.Logger
}

func NewProductHandler(svc service.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

// POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("create product failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}
	response.Created(c, p)
}

// GET /api/v1/products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

// PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
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

	p, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

// DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// GET /api/v1/products
func (h *ProductHandler) List(c *gin.Context) {
	var f domain.ProductFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, total, err := h.svc.List(c.Request.Context(), &f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, products, &response.Meta{
		Page:  f.Page,
		Limit: f.Limit,
		Total: total,
	})
}

// GET /api/v1/categories
func (h *ProductHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cats)
}
