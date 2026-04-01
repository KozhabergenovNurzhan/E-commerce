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

const OrderServiceName = "ecommerce.order.v1.OrderService"

// OrderHandler implements the OrderService gRPC service.
type OrderHandler struct {
	svc *service.Services
}

func NewOrderHandler(svc *service.Services) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderMessage, error) {
	userID, ok := interceptor.UserIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id in context")
	}

	items := make([]models.CreateOrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = models.CreateOrderItem{
			ProductID: it.ProductID,
			Quantity:  int(it.Quantity),
		}
	}
	co := &models.CreateOrder{Items: items}
	if req.AddressID != 0 {
		co.AddressID = &req.AddressID
	}

	order, err := h.svc.Order.Create(ctx, userID, co)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderMessage(order), nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderMessage, error) {
	order, err := h.svc.Order.GetByID(ctx, req.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderMessage(order), nil
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	userID, ok := interceptor.UserIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id in context")
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Limit)
	if limit < 1 {
		limit = 20
	}

	orders, total, err := h.svc.Order.ListByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, toGRPCError(err)
	}

	msgs := make([]*pb.OrderMessage, len(orders))
	for i, o := range orders {
		msgs[i] = toOrderMessage(o)
	}
	return &pb.ListOrdersResponse{
		Orders: msgs,
		Total:  int32(total),
		Page:   int32(page),
		Limit:  int32(limit),
	}, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	if err := h.svc.Order.UpdateStatus(ctx, req.ID, models.OrderStatus(req.Status)); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateOrderStatusResponse{}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	if err := h.svc.Order.Cancel(ctx, req.ID); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CancelOrderResponse{}, nil
}

// ── ServiceDesc ──────────────────────────────────────────────────────────────

type OrderServer interface {
	CreateOrder(context.Context, *pb.CreateOrderRequest) (*pb.OrderMessage, error)
	GetOrder(context.Context, *pb.GetOrderRequest) (*pb.OrderMessage, error)
	ListOrders(context.Context, *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error)
	UpdateOrderStatus(context.Context, *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error)
	CancelOrder(context.Context, *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error)
}

var OrderServiceDesc = grpc.ServiceDesc{
	ServiceName: OrderServiceName,
	HandlerType: (*OrderServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateOrder", Handler: orderCreateOrderHandler},
		{MethodName: "GetOrder", Handler: orderGetOrderHandler},
		{MethodName: "ListOrders", Handler: orderListOrdersHandler},
		{MethodName: "UpdateOrderStatus", Handler: orderUpdateOrderStatusHandler},
		{MethodName: "CancelOrder", Handler: orderCancelOrderHandler},
	},
	Streams: []grpc.StreamDesc{},
}

func orderCreateOrderHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.CreateOrderRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).CreateOrder(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + OrderServiceName + "/CreateOrder"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).CreateOrder(ctx, req.(*pb.CreateOrderRequest))
	})
}

func orderGetOrderHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.GetOrderRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).GetOrder(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + OrderServiceName + "/GetOrder"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).GetOrder(ctx, req.(*pb.GetOrderRequest))
	})
}

func orderListOrdersHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.ListOrdersRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).ListOrders(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + OrderServiceName + "/ListOrders"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).ListOrders(ctx, req.(*pb.ListOrdersRequest))
	})
}

func orderUpdateOrderStatusHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.UpdateOrderStatusRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).UpdateOrderStatus(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + OrderServiceName + "/UpdateOrderStatus"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).UpdateOrderStatus(ctx, req.(*pb.UpdateOrderStatusRequest))
	})
}

func orderCancelOrderHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := &pb.CancelOrderRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).CancelOrder(ctx, req)
	}
	return interceptor(ctx, req, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + OrderServiceName + "/CancelOrder"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).CancelOrder(ctx, req.(*pb.CancelOrderRequest))
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func toOrderMessage(o *models.Order) *pb.OrderMessage {
	msg := &pb.OrderMessage{
		ID:         o.ID,
		UserID:     o.UserID,
		Status:     string(o.Status),
		TotalPrice: o.TotalPrice,
		CreatedAt:  o.CreatedAt.Unix(),
		UpdatedAt:  o.UpdatedAt.Unix(),
	}
	if o.AddressID != nil {
		msg.AddressID = *o.AddressID
	}
	items := make([]*pb.OrderItemMessage, len(o.Items))
	for i, it := range o.Items {
		items[i] = &pb.OrderItemMessage{
			ID:        it.ID,
			OrderID:   it.OrderID,
			ProductID: it.ProductID,
			Quantity:  int32(it.Quantity),
			UnitPrice: it.UnitPrice,
		}
	}
	msg.Items = items
	return msg
}
