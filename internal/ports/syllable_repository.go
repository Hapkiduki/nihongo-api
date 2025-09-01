package ports

import (
	"context"
	"nihongo-api/internal/domain"
)

// SyllableRepository defines the interface for syllable data operations
type SyllableRepository interface {
	Create(ctx context.Context, syllable *domain.Syllable) error
	GetByID(ctx context.Context, id string) (*domain.Syllable, error)
	GetAll(ctx context.Context) ([]domain.Syllable, error)
	GetByType(ctx context.Context, syllableType domain.SyllableType) ([]domain.Syllable, error)
	Update(ctx context.Context, syllable *domain.Syllable) error
	Delete(ctx context.Context, id string) error
}
