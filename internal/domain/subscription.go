package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionExpired   SubscriptionStatus = "expired"
)

type Subscription struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	InternalUserID *primitive.ObjectID `bson:"internal_user_id,omitempty" json:"internal_user_id"`
	ExternalUserID string              `bson:"external_user_id" json:"external_user_id"` // RevenueCat app_user_id
	ProductID      string              `bson:"product_id" json:"product_id"`
	EventID        string              `bson:"event_id" json:"event_id"` // For idempotency
	Status         SubscriptionStatus  `bson:"status" json:"status"`
	ExpiresAt      time.Time           `bson:"expires_at" json:"expires_at"`
	EventType      string              `bson:"event_type" json:"event_type"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
}

// NewSubscription crea una nueva suscripción con valores por defecto
func NewSubscription(externalUserID, productID, eventID, eventType string, expiresAt time.Time) *Subscription {
	return &Subscription{
		ExternalUserID: externalUserID,
		ProductID:      productID,
		EventID:        eventID,
		EventType:      eventType,
		Status:         SubscriptionActive,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// UpdateStatus actualiza el estado y fecha de expiración
func (s *Subscription) UpdateStatus(status SubscriptionStatus, expiresAt time.Time) {
	s.Status = status
	s.UpdatedAt = time.Now()
	if !expiresAt.IsZero() {
		s.ExpiresAt = expiresAt
	}
}
