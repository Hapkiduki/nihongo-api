package mongo

import (
	"context"
	"errors"
	"time"

	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoSubscriptionRepository struct {
	collection *mongo.Collection
}

func NewMongoSubscriptionRepository(db *mongo.Database) ports.SubscriptionRepository {
	coll := db.Collection("subscriptions")
	// Create unique index for event_id idempotency
	indexEvent := mongo.IndexModel{
		Keys:    bson.D{{Key: "event_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_event_id"),
	}
	_, _ = coll.Indexes().CreateOne(context.Background(), indexEvent)

	// Create index for external_user_id for efficient queries
	indexExternal := mongo.IndexModel{
		Keys:    bson.D{{Key: "external_user_id", Value: 1}},
		Options: options.Index().SetName("idx_external_user_id"),
	}
	_, _ = coll.Indexes().CreateOne(context.Background(), indexExternal)

	return &mongoSubscriptionRepository{
		collection: coll,
	}
}

func (r *mongoSubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	if sub.ID.IsZero() {
		sub.ID = primitive.NewObjectID()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, sub)
	if mongo.IsDuplicateKeyError(err) {
		return ports.ErrDuplicateEvent // Idempotency
	}
	return err
}

func (r *mongoSubscriptionRepository) GetByEventID(ctx context.Context, eventID string) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.collection.FindOne(ctx, bson.M{"event_id": eventID}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Not found, safe for idempotency check
		}
		return nil, err
	}
	return &sub, nil
}

func (r *mongoSubscriptionRepository) UpdateByEventID(ctx context.Context, eventID string, sub *domain.Subscription) error {
	sub.UpdatedAt = time.Now()

	filter := bson.M{"event_id": eventID}
	update := bson.M{"$set": sub}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *mongoSubscriptionRepository) GetByExternalUserID(ctx context.Context, externalUserID string) ([]*domain.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"external_user_id": externalUserID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []*domain.Subscription
	if err = cursor.All(ctx, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *mongoSubscriptionRepository) UpdateInternalUserID(ctx context.Context, externalUserID string, internalUserID string) error {
	objID, err := primitive.ObjectIDFromHex(internalUserID)
	if err != nil {
		return err
	}

	filter := bson.M{"external_user_id": externalUserID, "internal_user_id": bson.M{"$exists": false}} // Solo huérfanas
	update := bson.M{
		"$set": bson.M{
			"internal_user_id": objID,
			"updated_at":       time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return nil // No changes, already linked or none
	}
	return nil
}
