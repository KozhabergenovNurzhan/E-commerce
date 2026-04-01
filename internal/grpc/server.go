// Package grpc wires together the gRPC server, interceptors, and handlers.
package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/handler"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/interceptor"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"

	// Register the JSON codec for the gRPC server (replaces protobuf default).
	_ "github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/codec"
)

// Server wraps a *grpc.Server with a pre-bound TCP listener.
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// New creates a gRPC server, registers all service handlers, and enables
// server reflection (useful for tools like grpcurl / grpcui).
func New(port string, svc *service.Services, authMgr auth.Manager) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("grpc: listen on port %s: %w", port, err)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryAuth(authMgr),
		),
	)

	// Register service handlers using their ServiceDesc (no protoc required).
	grpcSrv.RegisterService(&handler.AuthServiceDesc, handler.NewAuthHandler(svc))
	grpcSrv.RegisterService(&handler.ProductServiceDesc, handler.NewProductHandler(svc))
	grpcSrv.RegisterService(&handler.OrderServiceDesc, handler.NewOrderHandler(svc))

	// Enable server reflection so grpcurl / grpcui work out of the box.
	reflection.Register(grpcSrv)

	return &Server{grpcServer: grpcSrv, listener: lis}, nil
}

// Run starts serving and blocks until the server is stopped.
func (s *Server) Run() error {
	return s.grpcServer.Serve(s.listener)
}

// Stop performs a graceful shutdown, waiting for in-flight RPCs to finish.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
