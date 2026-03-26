package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/response"
)

const (
	CtxUserID   = "userID"
	CtxUserRole = "userRole"
)

// Auth validates the Bearer access token and sets userID + userRole in context.
func Auth(mgr auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing or invalid authorization header")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := mgr.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole allows only users with one of the given roles to proceed.
func RequireRole(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(CtxUserRole)
		if !exists {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		userRole, ok := val.(domain.Role)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "insufficient permissions")
		c.Abort()
	}
}

// MustUserID extracts the authenticated user's ID from context.
// Panics if called outside an Auth-protected route (programming error).
func MustUserID(c *gin.Context) int64 {
	return c.MustGet(CtxUserID).(int64)
}
