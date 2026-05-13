package grpc

import (
	"context"
	"errors"

	"payment_service/internal/models"
	"payment_service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	proto.UnimplementedPaymentServiceServer
	uc models.PaymentUsecase
}

func NewPaymentHandler(uc models.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.PaymentAccountResponse, error) {
	a, err := h.uc.CreateAccount(ctx, req.GetUserId(), req.GetCurrency())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.PaymentAccountResponse{Account: accountToProto(a)}, nil
}

func (h *PaymentHandler) GetAccount(ctx context.Context, req *proto.GetAccountRequest) (*proto.PaymentAccountResponse, error) {
	a, err := h.uc.GetAccount(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.PaymentAccountResponse{Account: accountToProto(a)}, nil
}

func (h *PaymentHandler) Deposit(ctx context.Context, req *proto.DepositRequest) (*proto.TransactionResponse, error) {
	t, a, err := h.uc.Deposit(ctx, req.GetUserId(), req.GetAmountCents(), req.GetCurrency(), req.GetProvider(), req.GetProviderReference())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.TransactionResponse{Transaction: transactionToProto(t), Account: accountToProto(a)}, nil
}

func (h *PaymentHandler) CreateEscrow(ctx context.Context, req *proto.CreateEscrowRequest) (*proto.EscrowResponse, error) {
	e, err := h.uc.CreateEscrow(ctx, req.GetJobId(), req.GetClientId(), req.GetFreelancerId(), req.GetAmountCents(), req.GetCurrency())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.EscrowResponse{Escrow: escrowToProto(e)}, nil
}

func (h *PaymentHandler) ReleaseEscrow(ctx context.Context, req *proto.ReleaseEscrowRequest) (*proto.EscrowResponse, error) {
	e, err := h.uc.ReleaseEscrow(ctx, req.GetEscrowId(), req.GetRequesterId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.EscrowResponse{Escrow: escrowToProto(e)}, nil
}

func (h *PaymentHandler) RefundEscrow(ctx context.Context, req *proto.RefundEscrowRequest) (*proto.EscrowResponse, error) {
	e, err := h.uc.RefundEscrow(ctx, req.GetEscrowId(), req.GetRequesterId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.EscrowResponse{Escrow: escrowToProto(e)}, nil
}

func (h *PaymentHandler) ListTransactions(ctx context.Context, req *proto.ListTransactionsRequest) (*proto.ListTransactionsResponse, error) {
	items, err := h.uc.ListTransactions(ctx, req.GetUserId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*proto.Transaction, 0, len(items))
	for i := range items {
		out = append(out, transactionToProto(&items[i]))
	}
	return &proto.ListTransactionsResponse{Transactions: out}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, models.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, models.ErrAccountNotFound), errors.Is(err, models.ErrEscrowNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, models.ErrInsufficientFunds), errors.Is(err, models.ErrEscrowNotHeld):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, models.ErrForbiddenRequester):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
