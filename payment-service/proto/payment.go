package proto

import (
	"context"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SendPaymentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SenderID   string  `protobuf:"bytes,1,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	ReceiverID string  `protobuf:"bytes,2,opt,name=receiver_id,json=receiverId,proto3" json:"receiver_id,omitempty"`
	Amount     float64 `protobuf:"fixed64,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Currency   string  `protobuf:"bytes,4,opt,name=currency,proto3" json:"currency,omitempty"`
	Note       string  `protobuf:"bytes,5,opt,name=note,proto3" json:"note,omitempty"`
}

type SendPaymentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionID string  `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Status        string  `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	SenderBalance float64 `protobuf:"fixed64,3,opt,name=sender_balance,json=senderBalance,proto3" json:"sender_balance,omitempty"`
	Message       string  `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	CreatedAt     string  `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
}

type GetTransactionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionID string `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
}

type GetTransactionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionID string  `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	SenderID      string  `protobuf:"bytes,2,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	ReceiverID    string  `protobuf:"bytes,3,opt,name=receiver_id,json=receiverId,proto3" json:"receiver_id,omitempty"`
	Amount        float64 `protobuf:"fixed64,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Currency      string  `protobuf:"bytes,5,opt,name=currency,proto3" json:"currency,omitempty"`
	Status        string  `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	Note          string  `protobuf:"bytes,7,opt,name=note,proto3" json:"note,omitempty"`
	CreatedAt     string  `protobuf:"bytes,8,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
}

type GetBalanceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type GetBalanceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID   string  `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Balance  float64 `protobuf:"fixed64,2,opt,name=balance,proto3" json:"balance,omitempty"`
	Currency string  `protobuf:"bytes,3,opt,name=currency,proto3" json:"currency,omitempty"`
}

func (x *SendPaymentRequest) Reset() {
	*x = SendPaymentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendPaymentRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*SendPaymentRequest) ProtoMessage()    {}
func (x *SendPaymentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *SendPaymentResponse) Reset() {
	*x = SendPaymentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendPaymentResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*SendPaymentResponse) ProtoMessage()    {}
func (x *SendPaymentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *GetTransactionRequest) Reset() {
	*x = GetTransactionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTransactionRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetTransactionRequest) ProtoMessage()    {}
func (x *GetTransactionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *GetTransactionResponse) Reset() {
	*x = GetTransactionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTransactionResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetTransactionResponse) ProtoMessage()    {}
func (x *GetTransactionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *GetBalanceRequest) Reset() {
	*x = GetBalanceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBalanceRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetBalanceRequest) ProtoMessage()    {}
func (x *GetBalanceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *GetBalanceResponse) Reset() {
	*x = GetBalanceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBalanceResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetBalanceResponse) ProtoMessage()    {}
func (x *GetBalanceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_payment_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

type PaymentServiceServer interface {
	SendPayment(context.Context, *SendPaymentRequest) (*SendPaymentResponse, error)
	GetTransaction(context.Context, *GetTransactionRequest) (*GetTransactionResponse, error)
	GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceResponse, error)
}

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
	if err := c.cc.Invoke(ctx, "/payment.PaymentService/SendPayment", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error) {
	out := new(GetTransactionResponse)
	if err := c.cc.Invoke(ctx, "/payment.PaymentService/GetTransaction", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) GetBalance(ctx context.Context, in *GetBalanceRequest, opts ...grpc.CallOption) (*GetBalanceResponse, error) {
	out := new(GetBalanceResponse)
	if err := c.cc.Invoke(ctx, "/payment.PaymentService/GetBalance", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/SendPayment"}
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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/GetTransaction"}
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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/GetBalance"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetBalance(ctx, req.(*GetBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "SendPayment", Handler: _PaymentService_SendPayment_Handler},
		{MethodName: "GetTransaction", Handler: _PaymentService_GetTransaction_Handler},
		{MethodName: "GetBalance", Handler: _PaymentService_GetBalance_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}

var file_payment_proto_msgTypes = make([]protoimpl.MessageInfo, 6)

func init() {
	file := buildPaymentProtoFileDescriptor()
	file_payment_proto_msgTypes[0].GoReflectType = reflect.TypeOf((*SendPaymentRequest)(nil))
	file_payment_proto_msgTypes[1].GoReflectType = reflect.TypeOf((*SendPaymentResponse)(nil))
	file_payment_proto_msgTypes[2].GoReflectType = reflect.TypeOf((*GetTransactionRequest)(nil))
	file_payment_proto_msgTypes[3].GoReflectType = reflect.TypeOf((*GetTransactionResponse)(nil))
	file_payment_proto_msgTypes[4].GoReflectType = reflect.TypeOf((*GetBalanceRequest)(nil))
	file_payment_proto_msgTypes[5].GoReflectType = reflect.TypeOf((*GetBalanceResponse)(nil))

	file_payment_proto_msgTypes[0].Desc = file.Messages().Get(0)
	file_payment_proto_msgTypes[1].Desc = file.Messages().Get(1)
	file_payment_proto_msgTypes[2].Desc = file.Messages().Get(2)
	file_payment_proto_msgTypes[3].Desc = file.Messages().Get(3)
	file_payment_proto_msgTypes[4].Desc = file.Messages().Get(4)
	file_payment_proto_msgTypes[5].Desc = file.Messages().Get(5)

	if !protoimpl.UnsafeEnabled {
		file_payment_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendPaymentRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_payment_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendPaymentResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_payment_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTransactionRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_payment_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTransactionResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_payment_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBalanceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_payment_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBalanceResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}

	filePaymentProto := buildPaymentProtoFileDescriptor()
	if err := protoregistry.GlobalFiles.RegisterFile(filePaymentProto); err != nil {
		panic(err)
	}
}

func buildPaymentProtoFileDescriptor() protoreflect.FileDescriptor {
	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("payment.proto"),
		Package: proto.String("payment"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("SendPaymentRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Number: proto.Int32(1), Name: proto.String("sender_id"), JsonName: proto.String("senderId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(2), Name: proto.String("receiver_id"), JsonName: proto.String("receiverId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(3), Name: proto.String("amount"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()},
					{Number: proto.Int32(4), Name: proto.String("currency"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(5), Name: proto.String("note"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
				},
			},
			{
				Name: proto.String("SendPaymentResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Number: proto.Int32(1), Name: proto.String("transaction_id"), JsonName: proto.String("transactionId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(2), Name: proto.String("status"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(3), Name: proto.String("sender_balance"), JsonName: proto.String("senderBalance"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()},
					{Number: proto.Int32(4), Name: proto.String("message"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(5), Name: proto.String("created_at"), JsonName: proto.String("createdAt"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
				},
			},
			{
				Name:  proto.String("GetTransactionRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{{Number: proto.Int32(1), Name: proto.String("transaction_id"), JsonName: proto.String("transactionId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()}},
			},
			{
				Name: proto.String("GetTransactionResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Number: proto.Int32(1), Name: proto.String("transaction_id"), JsonName: proto.String("transactionId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(2), Name: proto.String("sender_id"), JsonName: proto.String("senderId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(3), Name: proto.String("receiver_id"), JsonName: proto.String("receiverId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(4), Name: proto.String("amount"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()},
					{Number: proto.Int32(5), Name: proto.String("currency"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(6), Name: proto.String("status"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(7), Name: proto.String("note"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(8), Name: proto.String("created_at"), JsonName: proto.String("createdAt"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
				},
			},
			{
				Name:  proto.String("GetBalanceRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{{Number: proto.Int32(1), Name: proto.String("user_id"), JsonName: proto.String("userId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()}},
			},
			{
				Name: proto.String("GetBalanceResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Number: proto.Int32(1), Name: proto.String("user_id"), JsonName: proto.String("userId"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
					{Number: proto.Int32(2), Name: proto.String("balance"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()},
					{Number: proto.Int32(3), Name: proto.String("currency"), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("PaymentService"),
			Method: []*descriptorpb.MethodDescriptorProto{
				{Name: proto.String("SendPayment"), InputType: proto.String(".payment.SendPaymentRequest"), OutputType: proto.String(".payment.SendPaymentResponse")},
				{Name: proto.String("GetTransaction"), InputType: proto.String(".payment.GetTransactionRequest"), OutputType: proto.String(".payment.GetTransactionResponse")},
				{Name: proto.String("GetBalance"), InputType: proto.String(".payment.GetBalanceRequest"), OutputType: proto.String(".payment.GetBalanceResponse")},
			},
		}},
	}

	fd, err := protodesc.NewFile(file, nil)
	if err != nil {
		panic(err)
	}
	return fd
}
