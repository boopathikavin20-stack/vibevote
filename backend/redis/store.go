package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"pulsevote/models"
)

const (
	pollVoteKeyTemplate = "poll:%s:votes:%s"
	pollChannelTemplate = "poll:%s:updates"
)

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) IncrementVote(ctx context.Context, pollID, optionID string) (int64, error) {
	key := fmt.Sprintf(pollVoteKeyTemplate, pollID, optionID)
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) GetVoteCount(ctx context.Context, pollID, optionID string) (int64, error) {
	key := fmt.Sprintf(pollVoteKeyTemplate, pollID, optionID)
	return s.client.Get(ctx, key).Int64()
}

func (s *Store) GetPollTotals(ctx context.Context, pollID string) (map[string]int64, error) {
	keys := s.client.Keys(ctx, fmt.Sprintf("poll:%s:votes:*", pollID)).Val()
	result := make(map[string]int64)
	for _, key := range keys {
		optionID := key[stringsLastIndex(key, ":")+1:]
		count, err := s.client.Get(ctx, key).Int64()
		if err != nil {
			return nil, err
		}
		result[optionID] = count
	}
	return result, nil
}

func (s *Store) PublishVoteUpdate(ctx context.Context, event models.VoteUpdateEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.client.Publish(ctx, fmt.Sprintf(pollChannelTemplate, event.PollID), payload).Err()
}

func (s *Store) SetVoterKey(ctx context.Context, pollID, voterKey string, ttl time.Duration) error {
	return s.client.Set(ctx, fmt.Sprintf("poll:%s:voter:%s", pollID, voterKey), "1", ttl).Err()
}

func (s *Store) HasVoterKey(ctx context.Context, pollID, voterKey string) (bool, error) {
	val, err := s.client.Exists(ctx, fmt.Sprintf("poll:%s:voter:%s", pollID, voterKey)).Result()
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func stringsLastIndex(s string, sep string) int {
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			idx = i
		}
	}
	return idx
}
