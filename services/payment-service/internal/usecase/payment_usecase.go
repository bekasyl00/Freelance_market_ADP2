package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"payment_service/internal/models"
)

const (
	defaultCurrency = "USD"
	defaultLimit    = 20
	maxLimit        = 100
)

type paymentUsecase struct {
	repo models.PaymentRepository
	pub  models.EventPublisher
}

func NewPaymentUsecase(repo models.PaymentRepository, pub models.EventPublisher) models.PaymentUsecase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &paymentUsecase{repo: repo, pub: pub}
}

type noopPublisher struct{}

func (noopPublisher) PublishPaymentEvent(context.Context, models.PaymentEvent) error {
	return nil
}

func (u *paymentUsecase) CreateAccount(ctx context.Context, userID, currency string) (*models.PaymentAccount, error) {
	userID = strings.TrimSpace(userID)
	currency = normalizeCurrency(currency)
	if userID == "" {
		return nil, models.ErrInvalidInput
	}
	return u.repo.CreateAccount(ctx, userID, currency)
}

func (u *paymentUsecase) GetAccount(ctx context.Context, userID string) (*models.PaymentAccount, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, models.ErrInvalidInput
	}
	a, err := u.repo.GetAccountByUserID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAccountNotFound
	}
	return a, err
}

func (u *paymentUsecase) Deposit(ctx context.Context, userID string, amountCents int64, currency, provider, providerReference string) (*models.Transaction, *models.PaymentAccount, error) {
	userID = strings.TrimSpace(userID)
	currency = normalizeCurrency(currency)
	if userID == "" || amountCents <= 0 {
		return nil, nil, models.ErrInvalidInput
	}
	t, a, err := u.repo.Deposit(ctx, userID, amountCents, currency, strings.TrimSpace(provider), strings.TrimSpace(providerReference))
	if err != nil {
		return nil, nil, err
	}
	_ = u.pub.PublishPaymentEvent(ctx, models.PaymentEvent{
		EventType:     "payment.deposit.completed",
		UserID:        userID,
		TransactionID: t.ID,
		AmountCents:   amountCents,
		Currency:      currency,
		CreatedAtUnix: time.Now().Unix(),
	})
	return t, a, nil
}

func (u *paymentUsecase) CreateEscrow(ctx context.Context, jobID, clientID, freelancerID string, amountCents int64, currency string) (*models.Escrow, error) {
	jobID = strings.TrimSpace(jobID)
	clientID = strings.TrimSpace(clientID)
	freelancerID = strings.TrimSpace(freelancerID)
	currency = normalizeCurrency(currency)
	if jobID == "" || clientID == "" || freelancerID == "" || clientID == freelancerID || amountCents <= 0 {
		return nil, models.ErrInvalidInput
	}
	e, err := u.repo.CreateEscrow(ctx, jobID, clientID, freelancerID, amountCents, currency)
	if err != nil {
		return nil, err
	}
	_ = u.pub.PublishPaymentEvent(ctx, models.PaymentEvent{
		EventType:     "payment.escrow.held",
		UserID:        clientID,
		JobID:         jobID,
		EscrowID:      e.ID,
		AmountCents:   amountCents,
		Currency:      currency,
		CreatedAtUnix: time.Now().Unix(),
	})
	return e, nil
}

func (u *paymentUsecase) ReleaseEscrow(ctx context.Context, escrowID, requesterID string) (*models.Escrow, error) {
	escrowID = strings.TrimSpace(escrowID)
	requesterID = strings.TrimSpace(requesterID)
	if escrowID == "" || requesterID == "" {
		return nil, models.ErrInvalidInput
	}
	existing, err := u.repo.GetEscrowByID(ctx, escrowID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrEscrowNotFound
	}
	if err != nil {
		return nil, err
	}
	if existing.ClientID != requesterID {
		return nil, models.ErrForbiddenRequester
	}
	e, err := u.repo.ReleaseEscrow(ctx, escrowID)
	if err != nil {
		return nil, err
	}
	_ = u.pub.PublishPaymentEvent(ctx, models.PaymentEvent{
		EventType:     "payment.escrow.released",
		UserID:        e.FreelancerID,
		JobID:         e.JobID,
		EscrowID:      e.ID,
		AmountCents:   e.AmountCents,
		Currency:      e.Currency,
		CreatedAtUnix: time.Now().Unix(),
	})
	return e, nil
}

func (u *paymentUsecase) RefundEscrow(ctx context.Context, escrowID, requesterID string) (*models.Escrow, error) {
	escrowID = strings.TrimSpace(escrowID)
	requesterID = strings.TrimSpace(requesterID)
	if escrowID == "" || requesterID == "" {
		return nil, models.ErrInvalidInput
	}
	existing, err := u.repo.GetEscrowByID(ctx, escrowID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrEscrowNotFound
	}
	if err != nil {
		return nil, err
	}
	if existing.ClientID != requesterID {
		return nil, models.ErrForbiddenRequester
	}
	e, err := u.repo.RefundEscrow(ctx, escrowID)
	if err != nil {
		return nil, err
	}
	_ = u.pub.PublishPaymentEvent(ctx, models.PaymentEvent{
		EventType:     "payment.escrow.refunded",
		UserID:        e.ClientID,
		JobID:         e.JobID,
		EscrowID:      e.ID,
		AmountCents:   e.AmountCents,
		Currency:      e.Currency,
		CreatedAtUnix: time.Now().Unix(),
	})
	return e, nil
}

func (u *paymentUsecase) ListTransactions(ctx context.Context, userID string, limit int32, offset int64) ([]models.Transaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || offset < 0 {
		return nil, models.ErrInvalidInput
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return u.repo.ListTransactions(ctx, userID, limit, offset)
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return defaultCurrency
	}
	if len(currency) != 3 {
		return defaultCurrency
	}
	return currency
}
