package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetBalance(ctx context.Context, userID string) (float64, error) {
	var balance int64
	err := s.db.QueryRowContext(ctx, "SELECT balance_cents FROM users WHERE id = $1", userID).Scan(&balance)
	return float64(balance) / 100, err
}

func (s *Service) Transfer(ctx context.Context, fromID, toID, jobID string, amount float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cents := int64(amount * 100)

	// Check balance
	var balance int64
	if err := tx.QueryRowContext(ctx, "SELECT balance_cents FROM users WHERE id = $1 FOR UPDATE", fromID).Scan(&balance); err != nil {
		return err
	}
	if balance < cents {
		return errors.New("insufficient balance")
	}

	// Update balances
	if _, err := tx.ExecContext(ctx, "UPDATE users SET balance_cents = balance_cents - $1 WHERE id = $2", cents, fromID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE users SET balance_cents = balance_cents + $1 WHERE id = $2", cents, toID); err != nil {
		return err
	}

	// Record transaction
	if _, err := tx.ExecContext(ctx, "INSERT INTO transactions (user_id, amount_cents, type, description) VALUES ($1, $2, 'transfer', $3)", 
		fromID, -cents, fmt.Sprintf("Payment for job %s", jobID)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO transactions (user_id, amount_cents, type, description) VALUES ($1, $2, 'transfer', $3)", 
		toID, cents, fmt.Sprintf("Received payment for job %s", jobID)); err != nil {
		return err
	}

	return tx.Commit()
}


