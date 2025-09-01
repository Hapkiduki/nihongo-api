package ports

import (
	"context"
	"nihongo-api/internal/domain"
)

// CourseRepository defines the interface for course data operations
type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	GetByID(ctx context.Context, id string) (*domain.Course, error)
	GetAll(ctx context.Context) ([]domain.Course, error)
	GetByLevel(ctx context.Context, level domain.JLPTLevel) ([]domain.Course, error)
	GetPremium(ctx context.Context) ([]domain.Course, error)
	Update(ctx context.Context, course *domain.Course) error
	Delete(ctx context.Context, id string) error
}
