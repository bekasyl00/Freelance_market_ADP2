package domain

import "context"

type Role string

const (
	RoleAdmin      Role = "ADMIN"
	RoleClient     Role = "CLIENT"
	RoleFreelancer Role = "FREELANCER"
)

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         Role
	Skills       []string
	Rating       float32
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	UpdateUserSkills(ctx context.Context, id int64, skills []string) error
}

type UserUsecase interface {
	Register(ctx context.Context, username, email, password string, role Role) (int64, error) // Добавили role
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, id int64) (*User, error)
	UpdateSkills(ctx context.Context, id int64, skills []string) error
}
