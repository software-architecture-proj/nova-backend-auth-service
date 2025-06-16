package service

import (
	"context"
	"fmt"

	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	repo "github.com/software-architecture-proj/nova-backend-auth-service/internal/repository"
	mod "github.com/software-architecture-proj/nova-backend-auth-service/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repo.AuthServer
}

func NewAuthService(db *database.MongoDB) *AuthService {
	return &AuthService{
		repo: &repo.AuthServer{Db: db},
	}
}

func (s *AuthService) LogInWEmail(ctx context.Context, email, passw string) (*mod.UserV2, error) {
	us, err := s.repo.DBLogWEmail(ctx, email, passw)
	if err != nil {
		return nil, err
	}

	return us, nil
}

func (s *AuthService) NewUser(ctx context.Context, user *mod.UserV2) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	userID, err := s.repo.DBCreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	return userID, nil
}
