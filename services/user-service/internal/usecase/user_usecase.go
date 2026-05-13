package usecase

import (
	"context"
	"time"
	"user_service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your_secret_key_here")

type userUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{repo: repo}
}

func (u *userUsecase) Register(ctx context.Context, username, email, password string, role domain.Role) (int64, error) {
	switch role {
	case domain.RoleAdmin, domain.RoleClient, domain.RoleFreelancer:
	default:
		return 0, domain.ErrInvalidRole
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	return u.repo.CreateUser(ctx, user)
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, error) {

	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func (u *userUsecase) GetProfile(ctx context.Context, id int64) (*domain.User, error) {
	return u.repo.GetUserByID(ctx, id)
}

func (u *userUsecase) UpdateSkills(ctx context.Context, id int64, skills []string) error {

	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role != domain.RoleFreelancer {
		return domain.ErrOnlyFreelancerSkills
	}

	return u.repo.UpdateUserSkills(ctx, id, skills)
}
