package service

import (
	"context"
	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProgressService handles progress business logic
type ProgressService struct {
	progressRepo ports.ProgressRepository
}

// NewProgressService creates a new progress service
func NewProgressService(progressRepo ports.ProgressRepository) *ProgressService {
	return &ProgressService{
		progressRepo: progressRepo,
	}
}

// GetUserProgress retrieves all progress for a user
func (s *ProgressService) GetUserProgress(ctx context.Context, userID string) ([]domain.Progress, error) {
	return s.progressRepo.GetByUserID(ctx, userID)
}

// UpdateProgress updates or creates progress for a user
func (s *ProgressService) UpdateProgress(ctx context.Context, userID, entityID string, entityType domain.EntityType, completed bool, score int) error {
	progress, err := s.progressRepo.GetByUserAndEntity(ctx, userID, entityID, entityType)
	if err != nil {
		// Create new progress
		userObjID, _ := primitive.ObjectIDFromHex(userID)
		entityObjID, _ := primitive.ObjectIDFromHex(entityID)

		newProgress := &domain.Progress{
			UserID:     userObjID,
			EntityID:   entityObjID,
			EntityType: entityType,
			Completed:  completed,
			Score:      score,
		}

		if completed {
			now := time.Now()
			newProgress.CompletedAt = &now
		}

		return s.progressRepo.Create(ctx, newProgress)
	}

	// Update existing progress
	progress.Completed = completed
	progress.Score = score
	if completed && progress.CompletedAt == nil {
		now := time.Now()
		progress.CompletedAt = &now
	}

	return s.progressRepo.Update(ctx, progress)
}

// GetProgressByEntity retrieves progress for a specific entity
func (s *ProgressService) GetProgressByEntity(ctx context.Context, userID, entityID string, entityType domain.EntityType) (*domain.Progress, error) {
	return s.progressRepo.GetByUserAndEntity(ctx, userID, entityID, entityType)
}
