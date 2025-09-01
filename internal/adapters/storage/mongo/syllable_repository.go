package mongo

import (
	"context"
	"errors"
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
	return err
}

func (r *mongoSyllableRepository) GetByID(ctx context.Context, id string) (*domain.Syllable, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var syllable domain.Syllable
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&syllable)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("syllable not found")
		}
		return nil, err
	}
	return &syllable, nil
}

func (r *mongoSyllableRepository) GetAll(ctx context.Context) ([]domain.Syllable, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var syllables []domain.Syllable
	if err = cursor.All(ctx, &syllables); err != nil {
		return nil, err
	}
	return syllables, nil
}

func (r *mongoSyllableRepository) GetByType(ctx context.Context, syllableType domain.SyllableType) ([]domain.Syllable, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"type": syllableType})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var syllables []domain.Syllable
	if err = cursor.All(ctx, &syllables); err != nil {
		return nil, err
	}
	return syllables, nil
}

func (r *mongoSyllableRepository) Update(ctx context.Context, syllable *domain.Syllable) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": syllable.ID},
		bson.M{"$set": syllable},
	)
	return err
}

func (r *mongoSyllableRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
