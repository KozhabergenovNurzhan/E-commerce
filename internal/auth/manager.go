package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
)

// Claims is embedded in every access token.
type Claims struct {
	UserID int64       `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles stateless JWT operations (signing and parsing).
// Refresh token persistence is handled by the token service layer.
type Manager interface {
	GenerateAccessToken(userID int64, role domain.Role) (string, error)
	ValidateAccessToken(token string) (*Claims, error)
	AccessTTL() time.Duration
}
