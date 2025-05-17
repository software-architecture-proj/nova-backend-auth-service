package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	pb "github.com/software-architecture-proj/nova-backend-auth-service/services/genproto/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidUsernameLength = errors.New("username must be exactly 16 bytes")

// User represents the user model in the database
type User struct {
	ID        uuid.UUID `bson:"_id,omitempty" json:"id,omitempty" validate:"required"`
	Email     string    `bson:"email" json:"email" validate:"required,email"`
	Username  string    `bson:"username" json:"username" validate:"required,len=16"`
	Password  string    `bson:"password" json:"-" validate:"required,min=6"`
	FirstName string    `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string    `bson:"last_name" json:"last_name" validate:"required"`
	Phone     int64     `bson:"phone,omitempty" json:"phone,omitempty" validate:"omitempty,min=1000000000"`
	CodeID    int32     `bson:"code_id,omitempty" json:"code_id,omitempty" validate:"required"`
	Birthdate string    `bson:"birthdate" json:"birthdate"` // Format: "YYYY-MM-DD"
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// ToCreateRequest converts the User model to a CreateUserRequest message
func (u *User) ToCreateRequest() *pb.CreateUserRequest {
	return &pb.CreateUserRequest{
		Email:     u.Email,
		Username:  u.Username,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		CodeId:    u.CodeID,
		Birthdate: u.Birthdate,
	}
}

// FromCreateRequest creates a User model from a CreateUserRequest message
func UserFromCreateRequest(req *pb.CreateUserRequest) (*User, error) {
	now := time.Now()

	// Generate UUID
	id := uuid.New()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Validate username length
	if len(req.Username) != 16 {
		return nil, ErrInvalidUsernameLength
	}

	return &User{
		ID:        id,
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
	}, nil
}

// ComparePassword compares a plain text password with the hashed password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
