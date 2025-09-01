package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Lesson represents a lesson within a course
type Lesson struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Exercises []Exercise         `bson:"exercises" json:"exercises"`
}
