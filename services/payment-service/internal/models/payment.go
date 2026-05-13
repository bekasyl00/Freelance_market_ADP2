package models

import "context"

type EscrowStatus string

const (
	EscrowStatusHeld      EscrowStatus = "held"
	EscrowStatusReleased  EscrowStatus = "released"
	EscrowStatusRefunded  EscrowStatus = "refunded"
	EscrowStatusCancelled EscrowStatus = "cancelled"
)

type TransactionType string

const (
	TransactionTypeDeposit       TransactionType = "deposit"
	TransactionTypeEscrowHold    TransactionType = "escrow_hold"
	TransactionTypeEscrowRelease TransactionType = "escrow_release"
	TransactionTypeRefund        TransactionType = "refund"
	TransactionTypeWithdrawal    TransactionType = "withdrawal"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type PaymentAccount struct {
	ID             string
	UserID         string
	AvailableCents int64
	EscrowCents    int64
	Currency       string
	CreatedAtUnix  int64
	UpdatedAtUnix  int64
}

type Escrow struct {
	ID             string
	JobID          string
	ClientID       string
	FreelancerID   string
	AmountCents    int64
	Currency       string
	Status         EscrowStatus
	HeldAtUnix     int64
	ReleasedAtUnix int64
	RefundedAtUnix int64
}

type Transaction struct {
	ID                string
	UserID            string
	JobID             string
	EscrowID          string
	Type              TransactionType
	AmountCents       int64
	Currency          string
	Status            TransactionStatus
	Provider          string
	ProviderReference string
	CreatedAtUnix     int64
}

type PaymentEvent struct {
	EventType     string `json:"event_type"`
	UserID        string `json:"user_id"`
	JobID         string `json:"job_id,omitempty"`
	EscrowID      string `json:"escrow_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	AmountCents   int64  `json:"amount_cents"`
	Currency      string `json:"currency"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}

type PaymentRepository interface {
	CreateAccount(ctx context.Context, userID, currency string) (*PaymentAccount, error)
	GetAccountByUserID(ctx context.Context, userID string) (*PaymentAccount, error)
	Deposit(ctx context.Context, userID string, amountCents int64, currency, provider, providerReference string) (*Transaction, *PaymentAccount, error)
	CreateEscrow(ctx context.Context, jobID, clientID, freelancerID string, amountCents int64, currency string) (*Escrow, error)
	ReleaseEscrow(ctx context.Context, escrowID string) (*Escrow, error)
	RefundEscrow(ctx context.Context, escrowID string) (*Escrow, error)
	GetEscrowByID(ctx context.Context, escrowID string) (*Escrow, error)
	ListTransactions(ctx context.Context, userID string, limit int32, offset int64) ([]Transaction, error)
}

type EventPublisher interface {
	PublishPaymentEvent(ctx context.Context, e PaymentEvent) error
}

type PaymentUsecase interface {
	CreateAccount(ctx context.Context, userID, currency string) (*PaymentAccount, error)
	GetAccount(ctx context.Context, userID string) (*PaymentAccount, error)
	Deposit(ctx context.Context, userID string, amountCents int64, currency, provider, providerReference string) (*Transaction, *PaymentAccount, error)
	CreateEscrow(ctx context.Context, jobID, clientID, freelancerID string, amountCents int64, currency string) (*Escrow, error)
	ReleaseEscrow(ctx context.Context, escrowID, requesterID string) (*Escrow, error)
	RefundEscrow(ctx context.Context, escrowID, requesterID string) (*Escrow, error)
	ListTransactions(ctx context.Context, userID string, limit int32, offset int64) ([]Transaction, error)
}
