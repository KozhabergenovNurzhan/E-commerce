// Package interceptor provides gRPC server interceptors.
package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
)

type contextKey string

const (
	CtxUserID   contextKey = "userID"
	CtxUserRole contextKey = "userRole"
)

// publicMethods is the set of fully-qualified gRPC method paths that do NOT
// require a valid JWT access token.
var publicMethods = map[string]struct{}{
	"/ecommerce.auth.v1.AuthService/Register": {},
	"/ecommerce.auth.v1.AuthService/Login":    {},
	"/ecommerce.auth.v1.AuthService/Refresh":  {},
	"/ecommerce.auth.v1.AuthService/Logout":   {},

	// Product reads are public
	"/ecommerce.product.v1.ProductService/ListProducts":   {},
	"/ecommerce.product.v1.ProductService/GetProduct":     {},
	"/ecommerce.product.v1.ProductService/ListCategories": {},
}

// UnaryAuth returns a gRPC unary server interceptor that validates the Bearer
// JWT on all protected methods, and injects userID / userRole into the context.
func UnaryAuth(mgr auth.Manager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		ctx, err := authorize(ctx, mgr)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// authorize extracts the Bearer token from gRPC metadata, validates it, and
// returns an enriched context.
func authorize(ctx context.Context, mgr auth.Manager) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == values[0] {
		return nil, status.Error(codes.Unauthenticated, "authorization header must use Bearer scheme")
	}

	claims, err := mgr.ValidateAccessToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	ctx = context.WithValue(ctx, CtxUserID, claims.UserID)
	ctx = context.WithValue(ctx, CtxUserRole, claims.Role)
	return ctx, nil
}

// UserIDFromCtx extracts the authenticated user ID from the context.
func UserIDFromCtx(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(CtxUserID).(int64)
	return v, ok
}

// UserRoleFromCtx extracts the authenticated user role from the context.
func UserRoleFromCtx(ctx context.Context) (models.Role, bool) {
	v, ok := ctx.Value(CtxUserRole).(models.Role)
	return v, ok
}
