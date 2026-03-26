package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

// GET /api/v1/users
func (h *Handler) ListUsers(c *gin.Context) {
	var q struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=20"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	users, total, err := h.services.User.List(c.Request.Context(), q.Page, q.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, users, &response.Meta{Page: q.Page, Limit: q.Limit, Total: total})
}

// GET /api/v1/users/:id
func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// PUT /api/v1/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.services.User.Update(c.Request.Context(), id, req.FirstName, req.LastName)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// DELETE /api/v1/users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.services.User.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
