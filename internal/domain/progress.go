package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EntityType represents the type of entity being tracked
type EntityType string

const (
	SyllableEntity EntityType = "syllable"
	KanjiEntity    EntityType = "kanji"
	LessonEntity   EntityType = "lesson"
	ExerciseEntity EntityType = "exercise"
)

// Progress represents user progress on learning entities
type Progress struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	EntityID    primitive.ObjectID `bson:"entity_id" json:"entity_id"`
	EntityType  EntityType         `bson:"entity_type" json:"entity_type"`
	Completed   bool               `bson:"completed" json:"completed"`
	Score       int                `bson:"score" json:"score"` // For exercises
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}
