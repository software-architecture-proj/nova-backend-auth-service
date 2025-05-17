package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/software-architecture-proj/nova-backend-auth-service/config"
	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	pb "github.com/software-architecture-proj/nova-backend-auth-service/services/genproto/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	db *database.MongoDB
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Username  string             `bson:"username"`
	Password  string             `bson:"password"`
	FirstName string             `bson:"first_name"`
	LastName  string             `bson:"last_name"`
	Phone     string             `bson:"phone,omitempty"`
	CodeID    string             `bson:"code_id,omitempty"`
	Birthdate string             `bson:"birthdate"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (s *authServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request for email: %s", req.Email)

	// Create user document
	now := time.Now()
	user := &User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password, // Store password as-is for now
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		CodeID:    req.CodeId,
		Birthdate: req.Birthdate,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert user into database
	result, err := s.db.Users().InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return &pb.CreateUserResponse{
			Success: false,
			Message: "Failed to create user",
		}, fmt.Errorf("failed to create user: %v", err)
	}

	// Get the inserted ID
	insertedID := result.InsertedID.(primitive.ObjectID)
	log.Printf("Created user with ID: %s", insertedID.Hex())

	return &pb.CreateUserResponse{
		UserId:    insertedID.Hex(),
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

	// Find user by email and password
	var user User
	err := s.db.Users().FindOne(ctx, bson.M{
		"email":    req.Email,
		"password": req.Password,
	}).Decode(&user)

	if err != nil {
		log.Printf("Login failed: %v", err)
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, fmt.Errorf("login failed: %v", err)
	}

	log.Printf("User logged in successfully: %s", req.Email)

	return &pb.LoginResponse{
		Success:  true,
		Message:  "Login successful",
		UserId:   user.ID.Hex(),
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
