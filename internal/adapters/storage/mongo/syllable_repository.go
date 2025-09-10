package mongo

import (
	"context"
	"errors"
	"fmt"
	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoSyllableRepository implements ports.SyllableRepository
type mongoSyllableRepository struct {
	collection *mongo.Collection
}

// NewMongoSyllableRepository creates a new MongoDB syllable repository
func NewMongoSyllableRepository(db *mongo.Database) ports.SyllableRepository {
	return &mongoSyllableRepository{
		collection: db.Collection("syllables"),
	}
}

func (r *mongoSyllableRepository) Create(ctx context.Context, syllable *domain.Syllable) error {
	syllable.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, syllable)
	if err != nil {
		return fmt.Errorf("failed to create syllable: %w", err)
	}
	return nil
}

func (r *mongoSyllableRepository) GetByID(ctx context.Context, id string) (*domain.Syllable, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid syllable ID: %w", err)
	}

	var syllable domain.Syllable
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&syllable)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("syllable not found")
		}
		return nil, fmt.Errorf("failed to get syllable by ID: %w", err)
	}
	return &syllable, nil
}

func (r *mongoSyllableRepository) GetAll(ctx context.Context) ([]domain.Syllable, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all syllables: %w", err)
	}
	defer cursor.Close(ctx)

	var syllables []domain.Syllable
	if err = cursor.All(ctx, &syllables); err != nil {
		return nil, fmt.Errorf("failed to decode syllables: %w", err)
	}
	return syllables, nil
}

func (r *mongoSyllableRepository) GetByType(ctx context.Context, syllableType domain.SyllableType) ([]domain.Syllable, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"type": syllableType})
	if err != nil {
		return nil, fmt.Errorf("failed to get syllables by type: %w", err)
	}
	defer cursor.Close(ctx)

	var syllables []domain.Syllable
	if err = cursor.All(ctx, &syllables); err != nil {
		return nil, fmt.Errorf("failed to decode syllables: %w", err)
	}
	return syllables, nil
}

func (r *mongoSyllableRepository) Update(ctx context.Context, syllable *domain.Syllable) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": syllable.ID},
		bson.M{"$set": syllable},
	)
	if err != nil {
		return fmt.Errorf("failed to update syllable: %w", err)
	}
	return nil
}

func (r *mongoSyllableRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid syllable ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete syllable: %w", err)
	}
	return nil
}
