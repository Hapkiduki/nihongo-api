package ports

import (
	"context"
	"nihongo-api/internal/domain"
)

// KanjiRepository defines the interface for kanji data operations
type KanjiRepository interface {
	Create(ctx context.Context, kanji *domain.Kanji) error
	GetByID(ctx context.Context, id string) (*domain.Kanji, error)
	GetAll(ctx context.Context) ([]domain.Kanji, error)
	GetByLevel(ctx context.Context, level domain.JLPTLevel) ([]domain.Kanji, error)
	Update(ctx context.Context, kanji *domain.Kanji) error
	Delete(ctx context.Context, id string) error
}
