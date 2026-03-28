package auth

import (
	"fmt"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/golang-jwt/jwt/v5"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type JWTManager struct {
	secret    []byte
	accessTTL time.Duration
}

func NewJWTManager(secret string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) GenerateAccessToken(userID int64, role models.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        fmt.Sprintf("%d-%d", userID, time.Now().UnixNano()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) ValidateAccessToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.Unauthorized("unauthorized", nil)
		}
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperrors.Unauthorized("unauthorized", nil)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, apperrors.Unauthorized("unauthorized", nil)
	}
	return claims, nil
}

func (m *JWTManager) AccessTTL() time.Duration {
	return m.accessTTL
}
