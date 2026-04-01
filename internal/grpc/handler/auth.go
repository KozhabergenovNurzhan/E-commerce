// Package handler contains gRPC service handler implementations. Each handler
// is a thin translation layer: it converts pb types ↔ service models, delegates
// business logic to the service layer, and maps domain errors to gRPC status
// codes.
package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/pb"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/apperrors"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

// AuthServiceName is the fully-qualified gRPC service name as declared in the
// proto file. It is used when registering the ServiceDesc.
const AuthServiceName = "ecommerce.auth.v1.AuthService"

// AuthHandler implements the AuthService gRPC service.
type AuthHandler struct {
	svc *service.Services
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.Services) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register creates a new user account and returns auth tokens.
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	user, err := h.svc.User.Register(ctx, &models.Register{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	tokens, err := h.svc.Token.GenerateTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AuthResponse{
		User:   toUserMessage(user),
		Tokens: toTokensResponse(tokens),
	}, nil
}

// Login authenticates a user and returns auth tokens.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	record, err := h.svc.User.Login(ctx, &models.Login{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	tokens, err := h.svc.Token.GenerateTokenPair(ctx, record.ID, record.Role)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AuthResponse{
		User:   toUserMessage(record.ToResponse()),
		Tokens: toTokensResponse(tokens),
	}, nil
}

// Refresh exchanges a valid refresh token for a new token pair.
func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.TokensResponse, error) {
	tokens, err := h.svc.Token.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toTokensResponse(tokens), nil
}

// Logout revokes the provided refresh token.
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := h.svc.Token.Revoke(ctx, req.RefreshToken); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.LogoutResponse{}, nil
}

// ── ServiceDesc registration ─────────────────────────────────────────────────

// AuthServiceDesc is the grpc.ServiceDesc for AuthHandler.
var AuthServiceDesc = grpc.ServiceDesc{
	ServiceName: AuthServiceName,
	HandlerType: (*AuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    authRegisterHandler,
		},
		{
			MethodName: "Login",
			Handler:    authLoginHandler,
		},
		{
			MethodName: "Refresh",
			Handler:    authRefreshHandler,
		},
		{
			MethodName: "Logout",
			Handler:    authLogoutHandler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

// AuthServer is the interface the grpc.Server uses for type-checking.
type AuthServer interface {
	Register(context.Context, *pb.RegisterRequest) (*pb.AuthResponse, error)
	Login(context.Context, *pb.LoginRequest) (*pb.AuthResponse, error)
	Refresh(context.Context, *pb.RefreshRequest) (*pb.TokensResponse, error)
	Logout(context.Context, *pb.LogoutRequest) (*pb.LogoutResponse, error)
}

func authRegisterHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.RegisterRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Register(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + AuthServiceName + "/Register"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Register(ctx, req.(*pb.RegisterRequest))
	})
}

func authLoginHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.LoginRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Login(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + AuthServiceName + "/Login"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Login(ctx, req.(*pb.LoginRequest))
	})
}

func authRefreshHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.RefreshRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Refresh(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + AuthServiceName + "/Refresh"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Refresh(ctx, req.(*pb.RefreshRequest))
	})
}

func authLogoutHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.LogoutRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Logout(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + AuthServiceName + "/Logout"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Logout(ctx, req.(*pb.LogoutRequest))
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func toUserMessage(u *models.User) *pb.UserMessage {
	return &pb.UserMessage{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt.Unix(),
	}
}

func toTokensResponse(t *models.AuthTokens) *pb.TokensResponse {
	return &pb.TokensResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
	}
}

// toGRPCError converts a domain AppError (or generic error) to a gRPC status.
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.AppError
	if ae, ok := err.(*apperrors.AppError); ok {
		appErr = ae
	}
	if appErr == nil {
		return status.Error(codes.Internal, err.Error())
	}
	switch appErr.Code {
	case 400:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case 401:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case 403:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case 404:
		return status.Error(codes.NotFound, appErr.Message)
	case 409:
		return status.Error(codes.AlreadyExists, appErr.Message)
	case 422:
		return status.Error(codes.InvalidArgument, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
