package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/interceptor"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/grpc/pb"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/models"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

const ProductServiceName = "ecommerce.product.v1.ProductService"

// ProductHandler implements the ProductService gRPC service.
type ProductHandler struct {
	svc *service.Services
}

func NewProductHandler(svc *service.Services) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Limit)
	if limit < 1 {
		limit = 20
	}

	filter := &models.ProductFilter{
		Search: req.Search,
		Page:   page,
		Limit:  limit,
	}
	if req.CategoryID != 0 {
		filter.CategoryID = &req.CategoryID
	}
	if req.MinPrice != 0 {
		filter.MinPrice = &req.MinPrice
	}
	if req.MaxPrice != 0 {
		filter.MaxPrice = &req.MaxPrice
	}

	products, total, err := h.svc.Product.List(ctx, filter)
	if err != nil {
		return nil, toGRPCError(err)
	}

	msgs := make([]*pb.ProductMessage, len(products))
	for i, p := range products {
		msgs[i] = toProductMessage(p)
	}
	return &pb.ListProductsResponse{
		Products: msgs,
		Total:    int32(total),
		Page:     int32(page),
		Limit:    int32(limit),
	}, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductMessage, error) {
	p, err := h.svc.Product.GetByID(ctx, req.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProductMessage(p), nil
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductMessage, error) {
	userID, ok := interceptor.UserIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id in context")
	}
	role, _ := interceptor.UserRoleFromCtx(ctx)
	var sellerID *int64
	if role == models.RoleSeller {
		sellerID = &userID
	}

	p, err := h.svc.Product.Create(ctx, sellerID, &models.CreateProduct{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProductMessage(p), nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductMessage, error) {
	p, err := h.svc.Product.Update(ctx, req.ID, &models.UpdateProduct{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProductMessage(p), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if err := h.svc.Product.Delete(ctx, req.ID); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteProductResponse{}, nil
}

func (h *ProductHandler) ListCategories(ctx context.Context, _ *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	cats, err := h.svc.Product.ListCategories(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	msgs := make([]*pb.CategoryMessage, len(cats))
	for i, c := range cats {
		msgs[i] = &pb.CategoryMessage{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			CreatedAt: c.CreatedAt.Unix(),
		}
	}
	return &pb.ListCategoriesResponse{Categories: msgs}, nil
}

// ── ServiceDesc ──────────────────────────────────────────────────────────────

type ProductServer interface {
	ListProducts(context.Context, *pb.ListProductsRequest) (*pb.ListProductsResponse, error)
	GetProduct(context.Context, *pb.GetProductRequest) (*pb.ProductMessage, error)
	CreateProduct(context.Context, *pb.CreateProductRequest) (*pb.ProductMessage, error)
	UpdateProduct(context.Context, *pb.UpdateProductRequest) (*pb.ProductMessage, error)
	DeleteProduct(context.Context, *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error)
	ListCategories(context.Context, *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error)
}

var ProductServiceDesc = grpc.ServiceDesc{
	ServiceName: ProductServiceName,
	HandlerType: (*ProductServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ListProducts", Handler: productListProductsHandler},
		{MethodName: "GetProduct", Handler: productGetProductHandler},
		{MethodName: "CreateProduct", Handler: productCreateProductHandler},
		{MethodName: "UpdateProduct", Handler: productUpdateProductHandler},
		{MethodName: "DeleteProduct", Handler: productDeleteProductHandler},
		{MethodName: "ListCategories", Handler: productListCategoriesHandler},
	},
	Streams: []grpc.StreamDesc{},
}

func productListProductsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.ListProductsRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).ListProducts(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/ListProducts"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).ListProducts(ctx, req.(*pb.ListProductsRequest))
	})
}

func productGetProductHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.GetProductRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).GetProduct(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/GetProduct"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).GetProduct(ctx, req.(*pb.GetProductRequest))
	})
}

func productCreateProductHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.CreateProductRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).CreateProduct(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/CreateProduct"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).CreateProduct(ctx, req.(*pb.CreateProductRequest))
	})
}

func productUpdateProductHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.UpdateProductRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).UpdateProduct(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/UpdateProduct"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).UpdateProduct(ctx, req.(*pb.UpdateProductRequest))
	})
}

func productDeleteProductHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.DeleteProductRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).DeleteProduct(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/DeleteProduct"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).DeleteProduct(ctx, req.(*pb.DeleteProductRequest))
	})
}

func productListCategoriesHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.ListCategoriesRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServer).ListCategories(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + ProductServiceName + "/ListCategories"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServer).ListCategories(ctx, req.(*pb.ListCategoriesRequest))
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func toProductMessage(p *models.Product) *pb.ProductMessage {
	msg := &pb.ProductMessage{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       int32(p.Stock),
		ImageURL:    p.ImageURL,
		CreatedAt:   p.CreatedAt.Unix(),
	}
	if p.SellerID != nil {
		msg.SellerID = *p.SellerID
	}
	return msg
}
