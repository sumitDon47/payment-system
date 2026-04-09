package proto

import (
	"context"
	"google.golang.org/grpc"
)

// ─── Interface your server must implement ───────────────────────────────────

type PaymentServiceServer interface {
	SendPayment(context.Context, *SendPaymentRequest) (*SendPaymentResponse, error)
	GetTransaction(context.Context, *GetTransactionRequest) (*GetTransactionResponse, error)
	GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceResponse, error)
}

// ─── Request / Response structs ─────────────────────────────────────────────

type SendPaymentRequest struct {
	SenderID   string
	ReceiverID string
	Amount     float64
	Currency   string
	Note       string
}

type SendPaymentResponse struct {
	TransactionID string
	Status        string
	SenderBalance float64
	Message       string
	CreatedAt     string
}

type GetTransactionRequest struct {
	TransactionID string
}

type GetTransactionResponse struct {
	TransactionID string
	SenderID      string
	ReceiverID    string
	Amount        float64
	Currency      string
	Status        string
	Note          string
	CreatedAt     string
}

type GetBalanceRequest struct {
	UserID string
}

type GetBalanceResponse struct {
	UserID   string
	Balance  float64
	Currency string
}

// ─── Client stub (used by API Gateway to call this service) ─────────────────

type PaymentServiceClient interface {
	SendPayment(ctx context.Context, req *SendPaymentRequest, opts ...grpc.CallOption) (*SendPaymentResponse, error)
	GetTransaction(ctx context.Context, req *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error)
	GetBalance(ctx context.Context, req *GetBalanceRequest, opts ...grpc.CallOption) (*GetBalanceResponse, error)
}

type paymentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPaymentServiceClient(cc grpc.ClientConnInterface) PaymentServiceClient {
	return &paymentServiceClient{cc}
}

func (c *paymentServiceClient) SendPayment(ctx context.Context, in *SendPaymentRequest, opts ...grpc.CallOption) (*SendPaymentResponse, error) {
	out := new(SendPaymentResponse)
	err := c.cc.Invoke(ctx, "/payment.PaymentService/SendPayment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error) {
	out := new(GetTransactionResponse)
	err := c.cc.Invoke(ctx, "/payment.PaymentService/GetTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) GetBalance(ctx context.Context, in *GetBalanceRequest, opts ...grpc.CallOption) (*GetBalanceResponse, error) {
	out := new(GetBalanceResponse)
	err := c.cc.Invoke(ctx, "/payment.PaymentService/GetBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ─── Server Registration (used by your gRPC server) ─────────────────────────

func RegisterPaymentServiceServer(s grpc.ServiceRegistrar, srv PaymentServiceServer) {
	s.RegisterService(&PaymentService_ServiceDesc, srv)
}

func _PaymentService_SendPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendPaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).SendPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/SendPayment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).SendPayment(ctx, req.(*SendPaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_GetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).GetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/GetTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetTransaction(ctx, req.(*GetTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_GetBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).GetBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/GetBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetBalance(ctx, req.(*GetBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendPayment",
			Handler:    _PaymentService_SendPayment_Handler,
		},
		{
			MethodName: "GetTransaction",
			Handler:    _PaymentService_GetTransaction_Handler,
		},
		{
			MethodName: "GetBalance",
			Handler:    _PaymentService_GetBalance_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}
