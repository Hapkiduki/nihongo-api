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

type mongoCourseRepository struct {
	collection *mongo.Collection
}

func NewMongoCourseRepository(db *mongo.Database) ports.CourseRepository {
	return &mongoCourseRepository{
		collection: db.Collection("courses"),
	}
}

func (r *mongoCourseRepository) Create(ctx context.Context, course *domain.Course) error {
	course.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, course)
	if err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}
	return nil
}

func (r *mongoCourseRepository) GetByID(ctx context.Context, id string) (*domain.Course, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid course ID: %w", err)
	}

	var course domain.Course
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&course)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("course not found")
		}
		return nil, fmt.Errorf("failed to get course by ID: %w", err)
	}
	return &course, nil
}

func (r *mongoCourseRepository) GetAll(ctx context.Context) ([]domain.Course, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all courses: %w", err)
	}
	defer cursor.Close(ctx)

	var courses []domain.Course
	if err = cursor.All(ctx, &courses); err != nil {
		return nil, fmt.Errorf("failed to decode courses: %w", err)
	}
	return courses, nil
}

func (r *mongoCourseRepository) GetByLevel(ctx context.Context, level domain.JLPTLevel) ([]domain.Course, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"level": level})
	if err != nil {
		return nil, fmt.Errorf("failed to get courses by level: %w", err)
	}
	defer cursor.Close(ctx)

	var courses []domain.Course
	if err = cursor.All(ctx, &courses); err != nil {
		return nil, fmt.Errorf("failed to decode courses: %w", err)
	}
	return courses, nil
}

func (r *mongoCourseRepository) GetPremium(ctx context.Context) ([]domain.Course, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_premium": true})
	if err != nil {
		return nil, fmt.Errorf("failed to get premium courses: %w", err)
	}
	defer cursor.Close(ctx)

	var courses []domain.Course
	if err = cursor.All(ctx, &courses); err != nil {
		return nil, fmt.Errorf("failed to decode premium courses: %w", err)
	}
	return courses, nil
}

func (r *mongoCourseRepository) Update(ctx context.Context, course *domain.Course) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": course.ID},
		bson.M{"$set": course},
	)
	if err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}
	return nil
}

func (r *mongoCourseRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid course ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}
