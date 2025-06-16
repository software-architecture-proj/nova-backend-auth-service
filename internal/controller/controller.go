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

	// Create the producer
	producer, err := notification.NewProducer()
	if err != nil {
		log.Printf("Failed to create notification producer: %v", err)
		return badResponse(err.Error()), fmt.Errorf("validation error: %v", err)
	}
	defer producer.Close() // Always close the producer when done

	// Validate login request
	if err := validateLoginRequest(req); err != nil {
		return badResponse(err.Error()), fmt.Errorf("notif error: %v", err)
	}

	user, err := s.service.LogInWEmail(ctx, req.Email, req.Password)
	if err != nil {
		return badResponse(err.Error()), fmt.Errorf("failed to log in user: %v", err)
	}

	tokenString, err := buildToken(user)
	if err != nil {
		return badResponse(fmt.Sprintf("Failed to build token: %v", err)), fmt.Errorf("failed to build token: %v", err)
	}

	// Send the login notification
	err = producer.SendLoginNotification(req.Email)
	if err != nil {
		log.Printf("Failed to send login notification: %v", err)
		return badResponse(err.Error()), fmt.Errorf("notif error: %v", err)
	}
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
		return badResponse(err.Error()), fmt.Errorf("validation error: %v", err)
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
		return badResponse(err.Error()), fmt.Errorf("failed to create user: %v", err)
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
		"lastLog":  time.Now().Format("2006-01-02 15:04"),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"type":     "access",
		"httpOnly": true,
		"secure":   true,
	})

	// Sign both tokens
	accessTokenString, err := accessToken.SignedString([]byte("Thunderbolts*"))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %v", err)
	}

	return accessTokenString, nil
}

func parseJWT(tokenString string) (*jwt.Token, error) {
	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("Thunderbolts*"), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}
	return token, nil
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
