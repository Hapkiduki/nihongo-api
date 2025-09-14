package ports

import (
	"context"
	"errors"

	"nihongo-api/internal/domain"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByEventID(ctx context.Context, eventID string) (*domain.Subscription, error)
	UpdateByEventID(ctx context.Context, eventID string, sub *domain.Subscription) error
	GetByExternalUserID(ctx context.Context, externalUserID string) ([]*domain.Subscription, error)
	UpdateInternalUserID(ctx context.Context, externalUserID string, internalUserID string) error // Para sincronización
	// Opcional: GetActiveByInternalUserID(ctx context.Context, internalUserID string) (*domain.Subscription, error)
}

// Common errors for repositories
var (
	ErrDuplicateEvent = errors.New("subscription event already processed")
)
