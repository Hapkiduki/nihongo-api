package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Course represents a learning course
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" validate:"required,min=1,max=200"`
	Description string             `bson:"description" json:"description" validate:"required,min=1,max=1000"`
	Level       JLPTLevel          `bson:"level" json:"level" validate:"required,oneof=N5 N4 N3 N2 N1"`
	IsPremium   bool               `bson:"is_premium" json:"is_premium"`
	Lessons     []Lesson           `bson:"lessons" json:"lessons"`
}
