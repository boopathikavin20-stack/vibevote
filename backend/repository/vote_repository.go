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

type VoteRepository struct {
	collection *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	return &VoteRepository{collection: db.Collection("votes")}
}

func (r *VoteRepository) Create(ctx context.Context, vote *models.Vote) error {
	vote.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, vote)
	return err
}

func (r *VoteRepository) CountByPollAndOption(ctx context.Context, pollID, optionID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"pollId": pollID, "optionId": optionID})
}

func (r *VoteRepository) CountByPoll(ctx context.Context, pollID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"pollId": pollID})
}

func (r *VoteRepository) HasVoted(ctx context.Context, pollID, voterKey string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"pollId": pollID, "voterKey": voterKey})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *VoteRepository) HasUserVoted(ctx context.Context, pollID, userID string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"pollId": pollID, "voterId": userID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *VoteRepository) FindRecentByPoll(ctx context.Context, pollID string, limit int64) ([]models.Vote, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"pollId": pollID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var votes []models.Vote
	if err := cursor.All(ctx, &votes); err != nil {
		return nil, err
	}
	return votes, nil
}

func (r *VoteRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "pollId", Value: 1}, {Key: "optionId", Value: 1}},
	})
	if err != nil {
		return err
	}
	_, _ = r.collection.Indexes().DropOne(ctx, "pollId_1_voterKey_1")
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterKey", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		return err
	}
	_, _ = r.collection.Indexes().DropOne(ctx, "pollId_1_voterId_1")
	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterId", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *VoteRepository) FindOneByPollAndOption(ctx context.Context, pollID, optionID string) (*models.Vote, error) {
	var vote models.Vote
	err := r.collection.FindOne(ctx, bson.M{"pollId": pollID, "optionId": optionID}).Decode(&vote)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}
