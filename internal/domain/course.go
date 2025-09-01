package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Course represents a learning course
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Level       JLPTLevel          `bson:"level" json:"level"`
	IsPremium   bool               `bson:"is_premium" json:"is_premium"`
	Lessons     []Lesson           `bson:"lessons" json:"lessons"`
}
