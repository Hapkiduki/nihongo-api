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

type mongoProgressRepository struct {
	collection *mongo.Collection
}

func NewMongoProgressRepository(db *mongo.Database) ports.ProgressRepository {
	return &mongoProgressRepository{
		collection: db.Collection("progress"),
	}
}

func (r *mongoProgressRepository) Create(ctx context.Context, progress *domain.Progress) error {
	progress.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, progress)
	if err != nil {
		return fmt.Errorf("failed to create progress: %w", err)
	}
	return nil
}

func (r *mongoProgressRepository) GetByID(ctx context.Context, id string) (*domain.Progress, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid progress ID: %w", err)
	}

	var progress domain.Progress
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&progress)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("progress not found")
		}
		return nil, fmt.Errorf("failed to get progress by ID: %w", err)
	}
	return &progress, nil
}

func (r *mongoProgressRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Progress, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objID})
	if err != nil {
		return nil, fmt.Errorf("failed to get progress by user ID: %w", err)
	}
	defer cursor.Close(ctx)

	var progresses []domain.Progress
	if err = cursor.All(ctx, &progresses); err != nil {
		return nil, fmt.Errorf("failed to decode progresses: %w", err)
	}
	return progresses, nil
}

func (r *mongoProgressRepository) GetByUserAndEntity(ctx context.Context, userID, entityID string, entityType domain.EntityType) (*domain.Progress, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	entityObjID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	var progress domain.Progress
	err = r.collection.FindOne(ctx, bson.M{
		"user_id":     userObjID,
		"entity_id":   entityObjID,
		"entity_type": entityType,
	}).Decode(&progress)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("progress not found")
		}
		return nil, fmt.Errorf("failed to get progress by user and entity: %w", err)
	}
	return &progress, nil
}

func (r *mongoProgressRepository) Update(ctx context.Context, progress *domain.Progress) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": progress.ID},
		bson.M{"$set": progress},
	)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}
	return nil
}

func (r *mongoProgressRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid progress ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete progress: %w", err)
	}
	return nil
}
