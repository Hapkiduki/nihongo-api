package ports

import (
	"context"
	"nihongo-api/internal/domain"
)

// ProgressRepository defines the interface for progress data operations
type ProgressRepository interface {
	Create(ctx context.Context, progress *domain.Progress) error
	GetByID(ctx context.Context, id string) (*domain.Progress, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Progress, error)
	GetByUserAndEntity(ctx context.Context, userID, entityID string, entityType domain.EntityType) (*domain.Progress, error)
	Update(ctx context.Context, progress *domain.Progress) error
	Delete(ctx context.Context, id string) error
}
