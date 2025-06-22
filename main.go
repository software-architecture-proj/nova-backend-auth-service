package main

import (
	"log"

	server "github.com/software-architecture-proj/nova-backend-auth-service/internal/application"
)

func main() {
	log.Println("Starting Nova Auth Service...")
	server.InitializeServer("50053")
}
