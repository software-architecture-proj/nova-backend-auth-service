package application

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/software-architecture-proj/nova-backend-auth-service/config"
	"github.com/software-architecture-proj/nova-backend-auth-service/database"
	cont "github.com/software-architecture-proj/nova-backend-auth-service/internal/controller"
	pb "github.com/software-architecture-proj/nova-backend-common-protos/gen/go/auth_service"
	"google.golang.org/grpc"
)

func InitializeServer(port string) {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongodb, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB :50053")
	}
	defer mongodb.Close(context.Background())

	// Create gRPC server
	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &cont.AuthServer{Db: mongodb})

	// Start listening on a random available port
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen : %s", port)
	}
	log.Printf("Server is listening on port : %s", port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		server.GracefulStop()
	}()

	// Start server
	log.Printf("Starting gRPC server on port : %s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
