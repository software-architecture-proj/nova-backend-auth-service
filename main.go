package main

import (
	server "github.com/software-architecture-proj/nova-backend-auth-service/internal/application"
)

func main() {
	server.InitializeServer("50053")
}
