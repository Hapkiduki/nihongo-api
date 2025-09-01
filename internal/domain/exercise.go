package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExerciseType represents the type of exercise
type ExerciseType string

const (
	Quiz    ExerciseType = "quiz"
	Drawing ExerciseType = "drawing"
)

// Exercise represents an exercise within a lesson
type Exercise struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type     ExerciseType       `bson:"type" json:"type"`
	Question string             `bson:"question" json:"question"`
	Answer   string             `bson:"answer" json:"answer"`
	SVG      string             `bson:"svg,omitempty" json:"svg,omitempty"` // For drawing exercises
}
