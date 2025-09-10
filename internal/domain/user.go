package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name" validate:"required,min=2,max=100"`
	Email            string             `bson:"email" json:"email" validate:"required,email"`
	Password         string             `bson:"password" json:"-" validate:"required,min=8"` // Never expose password in JSON
	RevenueCatUserID string             `bson:"revenue_cat_user_id" json:"revenue_cat_user_id"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}
