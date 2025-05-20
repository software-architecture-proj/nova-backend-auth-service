FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN  go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o auth_service ./main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/auth_service .
CMD ["./auth_service"]

# To build the Docker image, run:
# docker build -t auth_service:x .