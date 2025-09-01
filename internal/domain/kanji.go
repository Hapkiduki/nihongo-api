package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JLPTLevel represents JLPT levels
type JLPTLevel string

const (
	N5 JLPTLevel = "N5"
	N4 JLPTLevel = "N4"
	N3 JLPTLevel = "N3"
	N2 JLPTLevel = "N2"
	N1 JLPTLevel = "N1"
)

// Kanji represents a kanji character
type Kanji struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Character string             `bson:"character" json:"character"`
	Meaning   string             `bson:"meaning" json:"meaning"`
	OnYomi    []string           `bson:"on_yomi" json:"on_yomi"`   // Chinese readings
	KunYomi   []string           `bson:"kun_yomi" json:"kun_yomi"` // Japanese readings
	SVG       string             `bson:"svg" json:"svg"`           // SVG for drawing strokes
	Level     JLPTLevel          `bson:"level" json:"level"`
}
