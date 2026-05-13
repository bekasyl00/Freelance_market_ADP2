package grpc

import (
	"payment_service/internal/models"
	"payment_service/proto"
)

func accountToProto(a *models.PaymentAccount) *proto.PaymentAccount {
	if a == nil {
		return nil
	}
	return &proto.PaymentAccount{
		Id:             a.ID,
		UserId:         a.UserID,
		AvailableCents: a.AvailableCents,
		EscrowCents:    a.EscrowCents,
		Currency:       a.Currency,
		CreatedAtUnix:  a.CreatedAtUnix,
		UpdatedAtUnix:  a.UpdatedAtUnix,
	}
}

func escrowToProto(e *models.Escrow) *proto.Escrow {
	if e == nil {
		return nil
	}
	return &proto.Escrow{
		Id:             e.ID,
		JobId:          e.JobID,
		ClientId:       e.ClientID,
		FreelancerId:   e.FreelancerID,
		AmountCents:    e.AmountCents,
		Currency:       e.Currency,
		Status:         escrowStatusToProto(e.Status),
		HeldAtUnix:     e.HeldAtUnix,
		ReleasedAtUnix: e.ReleasedAtUnix,
		RefundedAtUnix: e.RefundedAtUnix,
	}
}

func transactionToProto(t *models.Transaction) *proto.Transaction {
	if t == nil {
		return nil
	}
	return &proto.Transaction{
		Id:                t.ID,
		UserId:            t.UserID,
		JobId:             t.JobID,
		EscrowId:          t.EscrowID,
		Type:              transactionTypeToProto(t.Type),
		AmountCents:       t.AmountCents,
		Currency:          t.Currency,
		Status:            transactionStatusToProto(t.Status),
		Provider:          t.Provider,
		ProviderReference: t.ProviderReference,
		CreatedAtUnix:     t.CreatedAtUnix,
	}
}

func escrowStatusToProto(s models.EscrowStatus) proto.EscrowStatus {
	switch s {
	case models.EscrowStatusHeld:
		return proto.EscrowStatus_ESCROW_STATUS_HELD
	case models.EscrowStatusReleased:
		return proto.EscrowStatus_ESCROW_STATUS_RELEASED
	case models.EscrowStatusRefunded:
		return proto.EscrowStatus_ESCROW_STATUS_REFUNDED
	case models.EscrowStatusCancelled:
		return proto.EscrowStatus_ESCROW_STATUS_CANCELLED
	default:
		return proto.EscrowStatus_ESCROW_STATUS_UNSPECIFIED
	}
}

func transactionTypeToProto(t models.TransactionType) proto.TransactionType {
	switch t {
	case models.TransactionTypeDeposit:
		return proto.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case models.TransactionTypeEscrowHold:
		return proto.TransactionType_TRANSACTION_TYPE_ESCROW_HOLD
	case models.TransactionTypeEscrowRelease:
		return proto.TransactionType_TRANSACTION_TYPE_ESCROW_RELEASE
	case models.TransactionTypeRefund:
		return proto.TransactionType_TRANSACTION_TYPE_REFUND
	case models.TransactionTypeWithdrawal:
		return proto.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	default:
		return proto.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

func transactionStatusToProto(s models.TransactionStatus) proto.TransactionStatus {
	switch s {
	case models.TransactionStatusPending:
		return proto.TransactionStatus_TRANSACTION_STATUS_PENDING
	case models.TransactionStatusCompleted:
		return proto.TransactionStatus_TRANSACTION_STATUS_COMPLETED
	case models.TransactionStatusFailed:
		return proto.TransactionStatus_TRANSACTION_STATUS_FAILED
	case models.TransactionStatusCancelled:
		return proto.TransactionStatus_TRANSACTION_STATUS_CANCELLED
	default:
		return proto.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	}
}
