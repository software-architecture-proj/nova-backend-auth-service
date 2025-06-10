package service

import (
	"context"
	"fmt"

	repo "github.com/software-architecture-proj/nova-backend-auth-service/internal/repository"
	mod "github.com/software-architecture-proj/nova-backend-auth-service/models"
	"golang.org/x/crypto/bcrypt"
)

var Server = repo.AuthServer{}

func LogInWEmail(ctx context.Context, email, passw string) (*mod.UserV2, error) {
	us, err := Server.DBLogWEmail(ctx, email, passw)
	if err != nil {
		return nil, err
	}

	return us, nil
}

func NewUser(ctx context.Context, user *mod.UserV2) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	userID, err := Server.DBCreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	return userID, nil
}
