package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"pulsevote/models"
)

type PollRepository struct {
	collection *mongo.Collection
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	return &PollRepository{collection: db.Collection("polls")}
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.CreatedAt = time.Now()
	poll.UpdatedAt = poll.CreatedAt
	_, err := r.collection.InsertOne(ctx, poll)
	return err
}

func (r *PollRepository) FindByID(ctx context.Context, id string) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"shareCode": shareCode}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) ListByUser(ctx context.Context, userID string) ([]models.Poll, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"createdBy": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

func (r *PollRepository) Update(ctx context.Context, id string, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *PollRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *PollRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "shareCode", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdBy", Value: 1}},
	})
	return err
}
