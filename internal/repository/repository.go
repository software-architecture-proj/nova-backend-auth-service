package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	pb "github.com/software-architecture-proj/nova-backend-auth-service/gen/go/auth_service"
	mod "github.com/software-architecture-proj/nova-backend-auth-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	db *database.MongoDB
}

func (r *AuthServer) DBLogWEmail(ctx context.Context, email string, passw string) (*mod.UserV2, error) {
	var user mod.UserV2
	err := r.db.Users().FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email %s: %v", email, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passw)); err != nil {
		return nil, fmt.Errorf("invalid password for user %s", email)
	}
	log.Printf("User %s logged in successfully", email)
	return &user, nil
}

func (r *AuthServer) DBCreateUser(ctx context.Context, user *mod.UserV2) (string, error) {
	// Insert the new user
	_, err := r.db.Users().InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return "", fmt.Errorf("failed to insert user: %v", err)
	}

	log.Printf("Created new user with email: %s", user.Email)
	return user.ID.String(), nil
}
