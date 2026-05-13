package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"payment_service/internal/models"
)

type paymentPostgresRepo struct {
	db *sql.DB
}

func NewPaymentPostgresRepo(db *sql.DB) models.PaymentRepository {
	return &paymentPostgresRepo{db: db}
}

func (r *paymentPostgresRepo) CreateAccount(ctx context.Context, userID, currency string) (*models.PaymentAccount, error) {
	const q = `
INSERT INTO payment_accounts (user_id, currency)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET currency = payment_accounts.currency
RETURNING id, user_id, available_cents, escrow_cents, currency, created_at, updated_at`

	return scanAccountRow(r.db.QueryRowContext(ctx, q, userID, currency))
}

func (r *paymentPostgresRepo) GetAccountByUserID(ctx context.Context, userID string) (*models.PaymentAccount, error) {
	const q = `
SELECT id, user_id, available_cents, escrow_cents, currency, created_at, updated_at
FROM payment_accounts
WHERE user_id = $1`

	return scanAccountRow(r.db.QueryRowContext(ctx, q, userID))
}

func (r *paymentPostgresRepo) Deposit(ctx context.Context, userID string, amountCents int64, currency, provider, providerReference string) (*models.Transaction, *models.PaymentAccount, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer rollbackUnlessCommitted(tx)

	account, err := scanAccountRow(tx.QueryRowContext(ctx, `
INSERT INTO payment_accounts (user_id, available_cents, currency)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET available_cents = payment_accounts.available_cents + EXCLUDED.available_cents
RETURNING id, user_id, available_cents, escrow_cents, currency, created_at, updated_at`,
		userID, amountCents, currency,
	))
	if err != nil {
		return nil, nil, err
	}

	txn, err := scanTransactionRow(tx.QueryRowContext(ctx, `
INSERT INTO transactions (user_id, type, amount_cents, currency, status, provider, provider_reference)
VALUES ($1, 'deposit', $2, $3, 'completed', $4, $5)
RETURNING id, user_id, job_id, escrow_id, type, amount_cents, currency, status, provider, provider_reference, created_at`,
		userID, amountCents, currency, nullString(provider), nullString(providerReference),
	))
	if err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return txn, account, nil
}

func (r *paymentPostgresRepo) CreateEscrow(ctx context.Context, jobID, clientID, freelancerID string, amountCents int64, currency string) (*models.Escrow, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	res, err := tx.ExecContext(ctx, `
UPDATE payment_accounts
SET available_cents = available_cents - $1,
    escrow_cents = escrow_cents + $1
WHERE user_id = $2 AND available_cents >= $1`, amountCents, clientID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, models.ErrInsufficientFunds
	}

	escrow, err := scanEscrowRow(tx.QueryRowContext(ctx, `
INSERT INTO escrows (job_id, client_id, freelancer_id, amount_cents, currency, status)
VALUES ($1, $2, $3, $4, $5, 'held')
RETURNING id, job_id, client_id, freelancer_id, amount_cents, currency, status, held_at, released_at, refunded_at`,
		jobID, clientID, freelancerID, amountCents, currency,
	))
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO transactions (user_id, job_id, escrow_id, type, amount_cents, currency, status)
VALUES ($1, $2, $3, 'escrow_hold', $4, $5, 'completed')`,
		clientID, jobID, escrow.ID, amountCents, currency,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return escrow, nil
}

func (r *paymentPostgresRepo) ReleaseEscrow(ctx context.Context, escrowID string) (*models.Escrow, error) {
	return r.finishEscrow(ctx, escrowID, models.EscrowStatusReleased)
}

func (r *paymentPostgresRepo) RefundEscrow(ctx context.Context, escrowID string) (*models.Escrow, error) {
	return r.finishEscrow(ctx, escrowID, models.EscrowStatusRefunded)
}

func (r *paymentPostgresRepo) GetEscrowByID(ctx context.Context, escrowID string) (*models.Escrow, error) {
	const q = `
SELECT id, job_id, client_id, freelancer_id, amount_cents, currency, status, held_at, released_at, refunded_at
FROM escrows
WHERE id = $1`

	return scanEscrowRow(r.db.QueryRowContext(ctx, q, escrowID))
}

func (r *paymentPostgresRepo) ListTransactions(ctx context.Context, userID string, limit int32, offset int64) ([]models.Transaction, error) {
	const q = `
SELECT id, user_id, job_id, escrow_id, type, amount_cents, currency, status, provider, provider_reference, created_at
FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Transaction, 0)
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *paymentPostgresRepo) finishEscrow(ctx context.Context, escrowID string, next models.EscrowStatus) (*models.Escrow, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	escrow, err := scanEscrowRow(tx.QueryRowContext(ctx, `
SELECT id, job_id, client_id, freelancer_id, amount_cents, currency, status, held_at, released_at, refunded_at
FROM escrows
WHERE id = $1
FOR UPDATE`, escrowID))
	if err != nil {
		return nil, err
	}
	if escrow.Status != models.EscrowStatusHeld {
		return nil, models.ErrEscrowNotHeld
	}

	if next == models.EscrowStatusReleased {
		if _, err := tx.ExecContext(ctx, `
UPDATE payment_accounts
SET available_cents = available_cents + $1
WHERE user_id = $2`, escrow.AmountCents, escrow.FreelancerID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO transactions (user_id, job_id, escrow_id, type, amount_cents, currency, status)
VALUES ($1, $2, $3, 'escrow_release', $4, $5, 'completed')`,
			escrow.FreelancerID, escrow.JobID, escrow.ID, escrow.AmountCents, escrow.Currency); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
UPDATE payment_accounts
SET available_cents = available_cents + $1
WHERE user_id = $2`, escrow.AmountCents, escrow.ClientID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO transactions (user_id, job_id, escrow_id, type, amount_cents, currency, status)
VALUES ($1, $2, $3, 'refund', $4, $5, 'completed')`,
			escrow.ClientID, escrow.JobID, escrow.ID, escrow.AmountCents, escrow.Currency); err != nil {
			return nil, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE payment_accounts
SET escrow_cents = escrow_cents - $1
WHERE user_id = $2`, escrow.AmountCents, escrow.ClientID); err != nil {
		return nil, err
	}

	var q string
	if next == models.EscrowStatusReleased {
		q = `
UPDATE escrows
SET status = 'released', released_at = now()
WHERE id = $1
RETURNING id, job_id, client_id, freelancer_id, amount_cents, currency, status, held_at, released_at, refunded_at`
	} else {
		q = `
UPDATE escrows
SET status = 'refunded', refunded_at = now()
WHERE id = $1
RETURNING id, job_id, client_id, freelancer_id, amount_cents, currency, status, held_at, released_at, refunded_at`
	}
	updated, err := scanEscrowRow(tx.QueryRowContext(ctx, q, escrowID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return
	}
}

func scanAccountRow(row *sql.Row) (*models.PaymentAccount, error) {
	var created, updated time.Time
	a := &models.PaymentAccount{}
	if err := row.Scan(&a.ID, &a.UserID, &a.AvailableCents, &a.EscrowCents, &a.Currency, &created, &updated); err != nil {
		return nil, err
	}
	a.CreatedAtUnix = created.Unix()
	a.UpdatedAtUnix = updated.Unix()
	return a, nil
}

func scanEscrowRow(row *sql.Row) (*models.Escrow, error) {
	var held time.Time
	var released, refunded sql.NullTime
	e := &models.Escrow{}
	var status string
	if err := row.Scan(&e.ID, &e.JobID, &e.ClientID, &e.FreelancerID, &e.AmountCents, &e.Currency, &status, &held, &released, &refunded); err != nil {
		return nil, err
	}
	e.Status = models.EscrowStatus(status)
	e.HeldAtUnix = held.Unix()
	if released.Valid {
		e.ReleasedAtUnix = released.Time.Unix()
	}
	if refunded.Valid {
		e.RefundedAtUnix = refunded.Time.Unix()
	}
	return e, nil
}

func scanTransactionRow(row *sql.Row) (*models.Transaction, error) {
	var created time.Time
	var jobID, escrowID, provider, providerReference sql.NullString
	t := &models.Transaction{}
	var typ, status string
	if err := row.Scan(&t.ID, &t.UserID, &jobID, &escrowID, &typ, &t.AmountCents, &t.Currency, &status, &provider, &providerReference, &created); err != nil {
		return nil, err
	}
	fillTransaction(t, jobID, escrowID, typ, status, provider, providerReference, created)
	return t, nil
}

func scanTransaction(rows *sql.Rows) (*models.Transaction, error) {
	var created time.Time
	var jobID, escrowID, provider, providerReference sql.NullString
	t := &models.Transaction{}
	var typ, status string
	if err := rows.Scan(&t.ID, &t.UserID, &jobID, &escrowID, &typ, &t.AmountCents, &t.Currency, &status, &provider, &providerReference, &created); err != nil {
		return nil, err
	}
	fillTransaction(t, jobID, escrowID, typ, status, provider, providerReference, created)
	return t, nil
}

func fillTransaction(t *models.Transaction, jobID, escrowID sql.NullString, typ, status string, provider, providerReference sql.NullString, created time.Time) {
	if jobID.Valid {
		t.JobID = jobID.String
	}
	if escrowID.Valid {
		t.EscrowID = escrowID.String
	}
	if provider.Valid {
		t.Provider = provider.String
	}
	if providerReference.Valid {
		t.ProviderReference = providerReference.String
	}
	t.Type = models.TransactionType(typ)
	t.Status = models.TransactionStatus(status)
	t.CreatedAtUnix = created.Unix()
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}
