package repository

import (
	"context"
	"database/sql"
	"user_service/internal/domain"

	"github.com/lib/pq"
)

type userPostgresRepo struct {
	db *sql.DB
}

func NewUserPostgresRepo(db *sql.DB) domain.UserRepository {
	return &userPostgresRepo{db: db}
}

func (r *userPostgresRepo) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	// Добавили role в INSERT
	query := "INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int64
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash, string(user.Role)).Scan(&id)
	return id, err
}

func (r *userPostgresRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	var roleStr string
	// Добавили role в SELECT
	query := "SELECT id, username, email, password_hash, role, skills, rating FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &roleStr, pq.Array(&u.Skills), &u.Rating,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.Role(roleStr)
	return u, nil
}

func (r *userPostgresRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	u := &domain.User{}
	var roleStr string
	query := "SELECT id, username, email, role, skills, rating FROM users WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &roleStr, pq.Array(&u.Skills), &u.Rating,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.Role(roleStr)
	return u, nil
}

func (r *userPostgresRepo) UpdateUserSkills(ctx context.Context, id int64, skills []string) error {
	query := "UPDATE users SET skills = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, pq.Array(skills), id)
	return err
}
