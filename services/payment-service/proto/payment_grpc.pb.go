package proto

import (
	"context"

	"google.golang.org/grpc"
)

type PaymentServiceServer interface {
	CreateAccount(context.Context, *CreateAccountRequest) (*PaymentAccountResponse, error)
	GetAccount(context.Context, *GetAccountRequest) (*PaymentAccountResponse, error)
	Deposit(context.Context, *DepositRequest) (*TransactionResponse, error)
	CreateEscrow(context.Context, *CreateEscrowRequest) (*EscrowResponse, error)
	ReleaseEscrow(context.Context, *ReleaseEscrowRequest) (*EscrowResponse, error)
	RefundEscrow(context.Context, *RefundEscrowRequest) (*EscrowResponse, error)
	ListTransactions(context.Context, *ListTransactionsRequest) (*ListTransactionsResponse, error)
}

type UnimplementedPaymentServiceServer struct{}

func (UnimplementedPaymentServiceServer) CreateAccount(context.Context, *CreateAccountRequest) (*PaymentAccountResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) GetAccount(context.Context, *GetAccountRequest) (*PaymentAccountResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) Deposit(context.Context, *DepositRequest) (*TransactionResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) CreateEscrow(context.Context, *CreateEscrowRequest) (*EscrowResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) ReleaseEscrow(context.Context, *ReleaseEscrowRequest) (*EscrowResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) RefundEscrow(context.Context, *RefundEscrowRequest) (*EscrowResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) ListTransactions(context.Context, *ListTransactionsRequest) (*ListTransactionsResponse, error) {
	return nil, nil
}

func RegisterPaymentServiceServer(s grpc.ServiceRegistrar, srv PaymentServiceServer) {
	s.RegisterService(&PaymentService_ServiceDesc, srv)
}

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateAccount", Handler: _PaymentService_CreateAccount_Handler},
		{MethodName: "GetAccount", Handler: _PaymentService_GetAccount_Handler},
		{MethodName: "Deposit", Handler: _PaymentService_Deposit_Handler},
		{MethodName: "CreateEscrow", Handler: _PaymentService_CreateEscrow_Handler},
		{MethodName: "ReleaseEscrow", Handler: _PaymentService_ReleaseEscrow_Handler},
		{MethodName: "RefundEscrow", Handler: _PaymentService_RefundEscrow_Handler},
		{MethodName: "ListTransactions", Handler: _PaymentService_ListTransactions_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}

func _PaymentService_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/CreateAccount"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CreateAccount(ctx, req.(*CreateAccountRequest))
	})
}

func _PaymentService_GetAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).GetAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/GetAccount"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetAccount(ctx, req.(*GetAccountRequest))
	})
}

func _PaymentService_Deposit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DepositRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).Deposit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/Deposit"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).Deposit(ctx, req.(*DepositRequest))
	})
}

func _PaymentService_CreateEscrow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateEscrowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CreateEscrow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/CreateEscrow"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CreateEscrow(ctx, req.(*CreateEscrowRequest))
	})
}

func _PaymentService_ReleaseEscrow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReleaseEscrowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).ReleaseEscrow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/ReleaseEscrow"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).ReleaseEscrow(ctx, req.(*ReleaseEscrowRequest))
	})
}

func _PaymentService_RefundEscrow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefundEscrowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).RefundEscrow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/RefundEscrow"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).RefundEscrow(ctx, req.(*RefundEscrowRequest))
	})
}

func _PaymentService_ListTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).ListTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/ListTransactions"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).ListTransactions(ctx, req.(*ListTransactionsRequest))
	})
}
