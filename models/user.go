package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents the user model in the database
type User struct {
	ID        uuid.UUID `bson:"_id,omitempty" json:"id,omitempty" validate:"required"`
	Email     string    `bson:"email" json:"email" validate:"required,email"`
	Username  string    `bson:"username" json:"username" validate:"required,len=16"`
	Password  string    `bson:"password" json:"-" validate:"required,min=6"`
	FirstName string    `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string    `bson:"last_name" json:"last_name" validate:"required"`
	Phone     string    `bson:"phone,omitempty" json:"phone,omitempty" validate:"omitempty"`
	CodeID    string    `bson:"code_id,omitempty" json:"code_id,omitempty" validate:"required"`
	Birthdate string    `bson:"birthdate" json:"birthdate"` // Format: "YYYY-MM-DD"
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// ComparePassword compares a plain text password with the hashed password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
