package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/domain"
)

// Claims is embedded in every access token.
type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles stateless JWT operations (signing and parsing).
// Refresh token persistence is handled by the token service layer.
type Manager interface {
	GenerateAccessToken(userID uuid.UUID, role domain.Role) (string, error)
	ValidateAccessToken(token string) (*Claims, error)
	AccessTTL() time.Duration
}
