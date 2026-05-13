package models

import "context"

type JobApplicationSubmitted struct {
	JobID         int64  `json:"job_id"`
	FreelancerID  int64  `json:"freelancer_id"`
	ApplicationID int64  `json:"application_id"`
	CoverLetter   string `json:"cover_letter"`
	BidCents      int64  `json:"bid_cents"`
	CreatedAtUnix int64  `json:"created_at_unix"`
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

type Email struct {
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(ctx context.Context, email Email) error
}
