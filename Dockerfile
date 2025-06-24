# ./nova-backend-auth-service/Dockerfile.dev
# Development Dockerfile for the Auth service with Air live-reloading.

# Use the official Golang Alpine image.
FROM golang:1.24-alpine

# Set the working directory.
WORKDIR /app

# Install git, as it might be needed for Go modules.
RUN apk add --no-cache git

# Copy and download dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Install 'air' for live-reloading.
RUN go install github.com/air-verse/air@latest

# Copy the rest of the source code.
COPY . .

# Expose the service port for documentation.
EXPOSE 50053

# Start the service with air.
CMD ["air", "-c", ".air.toml"]
