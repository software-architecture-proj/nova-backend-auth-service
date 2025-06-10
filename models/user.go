package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user model in the database
type UserV2 struct {
	ID        uuid.UUID `bson:"_id,omitempty" json:"id,omitempty" validate:"required"`
	Email     string    `bson:"email" json:"email" validate:"required,email"`
	Phone     string    `bson:"phone,omitempty" json:"phone,omitempty" validate:"omitempty"`
	Password  string    `bson:"password" json:"-" validate:"required,min=6"`
	LastLog   string    `bson:"last_log" json:"last_log"` // Format: "YYYY-MM-DD HH:MM"
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
