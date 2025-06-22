package controller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	serv "github.com/software-architecture-proj/nova-backend-auth-service/internal/service"
	mod "github.com/software-architecture-proj/nova-backend-auth-service/models"
	"github.com/software-architecture-proj/nova-backend-auth-service/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/software-architecture-proj/nova-backend-common-protos/gen/go/auth_service"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	Db      *database.MongoDB
	service *serv.AuthService
}

func NewAuthServer(db *database.MongoDB) *AuthServer {
	return &AuthServer{
		Db:      db,
		service: serv.NewAuthService(db),
	}
}

func (s *AuthServer) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.Response, error) {
	log.Printf("Received login request for email: %s", req.Email)

	// Validate login request
	if err := validateLoginRequest(req); err != nil {
		return badResponse(err.Error()), status.Errorf(codes.Internal, "notif error: %v", err)
	}

	user, err := s.service.LogInWEmail(ctx, req.Email, req.Password)
	if err != nil {
		return badResponse(err.Error()), status.Errorf(codes.InvalidArgument, "failed to log in user: %v", err)
	}

	tokenString, err := buildToken(user)
	if err != nil {
		return badResponse(fmt.Sprintf("Failed to build token: %v", err)), status.Errorf(codes.Internal, "failed to build token: %v", err)
	}

	// Run notification in the background only if producer was created successfully
	go func() {
		producer, err := notification.NewProducer()
		if err != nil || producer == nil {
			log.Printf("Failed to create notification producer V2: %v", err)
			return
		}
		defer func() {
			if producer != nil {
				producer.Close()
			}
		}()
		if err := producer.SendLoginNotification(req.Email); err != nil {
			log.Printf("Failed to send login notification: %v", err)
		}
	}()

	log.Printf("User logged in successfully: %s", req.Email)
	return goodResponse("Login successful", tokenString), nil
}

func validateLoginRequest(req *pb.LoginRequest) error {
	// Validate email
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Basic email format validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	return nil
}

func (s *AuthServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.Response, error) {
	if err := validateSignUpRequest(req); err != nil {
		return badResponse(err.Error()), status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}
	now := time.Now()

	user := mod.UserV2{
		ID:        uuid.New(),
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Phone:     req.Phone,
		LastLog:   now.Format("2006-01-02 15:04"),
		CreatedAt: now,
		UpdatedAt: now,
	}
	userID, err := s.service.NewUser(ctx, &user)

	if err != nil {
		return badResponse(err.Error()), status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	return goodResponse("User created successfully", userID), nil
}
func validateSignUpRequest(req *pb.CreateUserRequest) error {
	// Validate email
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Basic email format validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	//Validate Phone
	if req.Phone != "" {
		if len(req.Phone) < 10 || len(req.Phone) > 15 {
			return fmt.Errorf("phone number must be between 10 and 15 digits")
		}

		for _, char := range req.Phone {
			if char < '0' || char > '9' {
				return fmt.Errorf("phone number must contain only digits")
			}

		}
	}
	return nil
}

func buildToken(user *mod.UserV2) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   user.ID,
		"email":    user.Email,
		"username": user.Username,
		"phone":    user.Phone,
		"lastLog":  time.Now().Format("2006-01-02 15:04"),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"type":     "access",
		"httpOnly": true,
		"secure":   true,
		"sameSite": "None",
	})

	// Sign both tokens
	accessTokenString, err := accessToken.SignedString([]byte("Thunderbolts*"))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %v", err)
	}

	return accessTokenString, nil
}

// Responses
func badResponse(message string) *pb.Response {
	return &pb.Response{
		Success: false,
		Message: message,
	}
}

func goodResponse(message string, data string) *pb.Response {
	return &pb.Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}
