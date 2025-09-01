package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyllableType represents the type of syllable (hiragana or katakana)
type SyllableType string

const (
	Hiragana SyllableType = "hiragana"
	Katakana SyllableType = "katakana"
)

// Syllable represents a hiragana or katakana syllable
type Syllable struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Symbol  string             `bson:"symbol" json:"symbol"`
	Reading string             `bson:"reading" json:"reading"`
	Type    SyllableType       `bson:"type" json:"type"`
	SVG     string             `bson:"svg" json:"svg"` // SVG for drawing strokes
}
