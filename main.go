package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/software-architecture-proj/nova-backend-auth-service/config"
	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	pb "github.com/software-architecture-proj/nova-backend-auth-service/services/genproto/auth"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	db *database.MongoDB
}

type User struct {
	ID        uuid.UUID `bson:"_id,omitempty"`
	Email     string    `bson:"email"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	FirstName string    `bson:"first_name"`
	LastName  string    `bson:"last_name"`
	Phone     string    `bson:"phone,omitempty"`
	CodeID    string    `bson:"code_id,omitempty"`
	Birthdate string    `bson:"birthdate"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (s *authServer) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Received login request for email: %s", req.Email)

	// Validate login request
	if err := validateLoginRequest(req); err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Find user by email
	var user User
	err := s.db.Users().FindOne(ctx, bson.M{
		"email": req.Email,
	}).Decode(&user)

	if err != nil {
		log.Printf("User not found: %v", err)
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, fmt.Errorf("user not found: %v", err)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Invalid password for user %s", req.Email)
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, fmt.Errorf("invalid password")
	}

	log.Printf("User logged in successfully: %s", req.Email)

	response := &pb.LoginResponse{
		Success: true,
		Message: "Login successful",
		Email:   user.Email,
	}

	// Debug log for final response
	log.Printf("Debug - Final Response: %+v", response)

	return response, nil
}

// validateLoginRequest validates the login request fields
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

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongodb, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB :50051")
	}
	defer mongodb.Close(context.Background())

	// Create gRPC server
	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &authServer{db: mongodb})

	// Start listening on a random available port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen :50051")
	}
	log.Printf("Server is listening on port :50051")

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		server.GracefulStop()
	}()

	// Start server
	log.Printf("Starting gRPC server on port :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
