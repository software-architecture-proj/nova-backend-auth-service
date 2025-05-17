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
	"go.mongodb.org/mongo-driver/mongo"
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
	Phone     int64     `bson:"phone,omitempty"`
	CodeID    int32     `bson:"code_id,omitempty"`
	Birthdate string    `bson:"birthdate"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (s *authServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request for email: %s", req.Email)

	// Validate username length
	if len(req.Username) != 16 {
		return &pb.CreateUserResponse{
			Success: false,
			Message: fmt.Sprintf("Username must be exactly 16 characters long. Your username has %d characters.", len(req.Username)),
		}, fmt.Errorf("invalid username length: got %d, want 16", len(req.Username))
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.CreateUserResponse{
			Success: false,
			Message: "There was a problem processing your password. Please try again with a different password.",
		}, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user document
	now := time.Now()
	user := &User{
		ID:        uuid.New(),
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		CodeID:    req.CodeId,
		Birthdate: req.Birthdate,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert user into database
	_, err = s.db.Users().InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				return &pb.CreateUserResponse{
					Success: false,
					Message: "This email is already registered. Please use a different email address.",
				}, fmt.Errorf("duplicate email: %v", err)
			}
			if strings.Contains(err.Error(), "username") {
				return &pb.CreateUserResponse{
					Success: false,
					Message: "This username is already taken. Please choose a different username.",
				}, fmt.Errorf("duplicate username: %v", err)
			}
			return &pb.CreateUserResponse{
				Success: false,
				Message: "A user with these details already exists.",
			}, fmt.Errorf("duplicate key: %v", err)
		}
		log.Printf("Failed to create user: %v", err)
		return &pb.CreateUserResponse{
			Success: false,
			Message: "An error occurred while creating your account. Please try again later.",
		}, fmt.Errorf("failed to create user: %v", err)
	}

	// Get the inserted ID
	insertedID := user.ID.String()
	log.Printf("Created user with ID: %s", insertedID)

	return &pb.CreateUserResponse{
		UserId:    insertedID,
		Success:   true,
		Message:   "User created successfully",
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

func (s *authServer) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Received login request for email: %s", req.Email)

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

	return &pb.LoginResponse{
		Success:  true,
		Message:  "Login successful",
		UserId:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
	}, nil
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongodb, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close(context.Background())

	// Create gRPC server
	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &authServer{db: mongodb})

	// Start listening on a random available port
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	log.Printf("Server is listening on port :%d", port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		server.GracefulStop()
	}()

	// Start server
	log.Printf("Starting gRPC server on port :%d", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
