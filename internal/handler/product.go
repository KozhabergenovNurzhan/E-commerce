package handler

import (
	"log/slog"
	"strconv"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

// GET /api/v1/products
func (h *Handler) ListProducts(c *gin.Context) {
	var f models.ProductFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if f.MinPrice != nil && f.MaxPrice != nil && *f.MinPrice > *f.MaxPrice {
		response.BadRequest(c, "min_price cannot be greater than max_price")
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	var req models.CreateProduct
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var sellerID *int64
	callerRole := c.MustGet(middleware.CtxUserRole).(models.Role)
	if callerRole == models.RoleSeller {
		id := middleware.MustUserID(c)
		sellerID = &id
	}

	p, err := h.services.Product.Create(c.Request.Context(), sellerID, &req)
	if err != nil {
		h.logger.Error("create product failed", slog.String("err", err.Error()))
		response.Error(c, err)
		return
	}
	response.Created(c, p)
}

// PUT /api/v1/products/:id
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	callerID := middleware.MustUserID(c)
	callerRole := c.MustGet(middleware.CtxUserRole).(models.Role)
	if callerRole == models.RoleSeller {
		p, err := h.services.Product.GetByID(c.Request.Context(), id)
		if err != nil {
			response.Error(c, err)
			return
		}
		if p.SellerID == nil || *p.SellerID != callerID {
			response.Forbidden(c, "cannot update another seller's product")
			return
		}
	}

	var req models.UpdateProduct
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	callerID := middleware.MustUserID(c)
	callerRole := c.MustGet(middleware.CtxUserRole).(models.Role)
	if callerRole == models.RoleSeller {
		p, err := h.services.Product.GetByID(c.Request.Context(), id)
		if err != nil {
			response.Error(c, err)
			return
		}
		if p.SellerID == nil || *p.SellerID != callerID {
			response.Forbidden(c, "cannot delete another seller's product")
			return
		}
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

// GET /api/v1/seller/products
func (h *Handler) ListSellerProducts(c *gin.Context) {
	sellerID := middleware.MustUserID(c)

	var f models.ProductFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, total, err := h.services.Product.ListBySeller(c.Request.Context(), sellerID, &f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, products, &response.Meta{Page: f.Page, Limit: f.Limit, Total: total})
}
