package models

import (
	"time"

	pb "github.com/software-architecture-proj/nova-backend-auth-service/services/genproto/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents the user model in the database
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" validate:"required"`
	Email     string             `bson:"email" json:"email" validate:"required,email"`
	Username  string             `bson:"username" json:"username" validate:"required,min=3"`
	Password  string             `bson:"password" json:"-" validate:"required,min=6"`
	FirstName string             `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string             `bson:"last_name" json:"last_name" validate:"required"`
	Phone     string             `bson:"phone,omitempty" json:"phone,omitempty"`
	CodeID    string             `bson:"code_id,omitempty" json:"code_id,omitempty"`
	Birthdate string             `bson:"birthdate" json:"birthdate"` // Format: "YYYY-MM-DD"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
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
func UserFromCreateRequest(req *pb.CreateUserRequest) *User {
	now := time.Now()
	return &User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		CodeID:    req.CodeId,
		Birthdate: req.Birthdate,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
