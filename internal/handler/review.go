package handler

import (
	"strconv"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/middleware"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// GET /api/v1/products/:id/reviews
func (h *Handler) ListReviews(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var q struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=20"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	reviews, total, rating, err := h.services.Review.ListByProduct(c.Request.Context(), productID, q.Page, q.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    reviews,
		"rating":  rating,
		"meta":    response.Meta{Page: q.Page, Limit: q.Limit, Total: total},
	})
}

// POST /api/v1/products/:id/reviews
func (h *Handler) CreateReview(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req models.CreateReview
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.MustUserID(c)

	review, err := h.services.Review.Create(c.Request.Context(), userID, productID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, review)
}

// PUT /api/v1/products/:id/reviews/:reviewId
func (h *Handler) UpdateReview(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("reviewId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid review id")
		return
	}

	var req models.UpdateReview
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.MustUserID(c)

	review, err := h.services.Review.Update(c.Request.Context(), reviewID, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, review)
}

// DELETE /api/v1/products/:id/reviews/:reviewId
func (h *Handler) DeleteReview(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("reviewId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid review id")
		return
	}

	callerID := middleware.MustUserID(c)
	callerRole := c.MustGet(middleware.CtxUserRole).(models.Role)

	if err := h.services.Review.Delete(c.Request.Context(), reviewID, callerID, callerRole); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
