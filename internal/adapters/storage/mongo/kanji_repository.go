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

type mongoKanjiRepository struct {
	collection *mongo.Collection
}

func NewMongoKanjiRepository(db *mongo.Database) ports.KanjiRepository {
	return &mongoKanjiRepository{
		collection: db.Collection("kanji"),
	}
}

func (r *mongoKanjiRepository) Create(ctx context.Context, kanji *domain.Kanji) error {
	kanji.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, kanji)
	if err != nil {
		return fmt.Errorf("failed to create kanji: %w", err)
	}
	return nil
}

func (r *mongoKanjiRepository) GetByID(ctx context.Context, id string) (*domain.Kanji, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid kanji ID: %w", err)
	}

	var kanji domain.Kanji
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&kanji)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("kanji not found")
		}
		return nil, fmt.Errorf("failed to get kanji by ID: %w", err)
	}
	return &kanji, nil
}

func (r *mongoKanjiRepository) GetAll(ctx context.Context) ([]domain.Kanji, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all kanji: %w", err)
	}
	defer cursor.Close(ctx)

	var kanjiList []domain.Kanji
	if err = cursor.All(ctx, &kanjiList); err != nil {
		return nil, fmt.Errorf("failed to decode kanji: %w", err)
	}
	return kanjiList, nil
}

func (r *mongoKanjiRepository) GetByLevel(ctx context.Context, level domain.JLPTLevel) ([]domain.Kanji, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"level": level})
	if err != nil {
		return nil, fmt.Errorf("failed to get kanji by level: %w", err)
	}
	defer cursor.Close(ctx)

	var kanjiList []domain.Kanji
	if err = cursor.All(ctx, &kanjiList); err != nil {
		return nil, fmt.Errorf("failed to decode kanji: %w", err)
	}
	return kanjiList, nil
}

func (r *mongoKanjiRepository) Update(ctx context.Context, kanji *domain.Kanji) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": kanji.ID},
		bson.M{"$set": kanji},
	)
	if err != nil {
		return fmt.Errorf("failed to update kanji: %w", err)
	}
	return nil
}

func (r *mongoKanjiRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid kanji ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete kanji: %w", err)
	}
	return nil
}
