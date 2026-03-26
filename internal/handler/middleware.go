package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

const (
	ctxUserID   = "userID"
	ctxUserRole = "userRole"
)

// JWTMiddleware validates the Bearer token and sets userID + userRole in context.
func JWTMiddleware(tokenSvc service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing or invalid authorization header")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUserRole, claims.Role)
		c.Next()
	}
}

// mustUserID extracts the authenticated user's UUID from context.
// Panics if called outside a JWT-protected route (programming error).
func mustUserID(c *gin.Context) uuid.UUID {
	return c.MustGet(ctxUserID).(uuid.UUID)
}
