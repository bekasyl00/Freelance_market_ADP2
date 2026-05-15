package auth

import (
	"context"
	"database/sql"
	"errors"
	// "golang.org/x/crypto/bcrypt"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, email, password, fullName, role string) (string, error) {
	// hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hash := password // Temporary plain text for offline dev
	var id string
	err := s.db.QueryRowContext(ctx, 
		"INSERT INTO users (email, password_hash, full_name, role) VALUES ($1, $2, $3, $4) RETURNING id",
		email, string(hash), fullName, role).Scan(&id)
	return id, err
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	var id, hash, role string
	err := s.db.QueryRowContext(ctx, "SELECT id, password_hash, role FROM users WHERE email = $1", email).Scan(&id, &hash, &role)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	// if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
	if hash != password {
		return "", "", errors.New("invalid password")
	}
	return id, role, nil
}

func (s *Service) GetProfile(ctx context.Context, userID string) (map[string]any, error) {
	var name, role string
	var rating float64
	var completed int
	err := s.db.QueryRowContext(ctx, "SELECT full_name, role, rating, completed_jobs FROM users WHERE id = $1", userID).Scan(&name, &role, &rating, &completed)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": userID,
		"name": name,
		"role": role,
		"rating": rating,
		"completedJobs": completed,
	}, nil
}
