package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"pulsevote/models"
	"pulsevote/redis"
	"pulsevote/repository"
)

type PollService struct {
	pollRepo *repository.PollRepository
	voteRepo *repository.VoteRepository
	redis    *redis.Store
}

func NewPollService(pollRepo *repository.PollRepository, voteRepo *repository.VoteRepository, store *redis.Store) *PollService {
	return &PollService{pollRepo: pollRepo, voteRepo: voteRepo, redis: store}
}

func (s *PollService) CreatePoll(ctx context.Context, userID string, req models.CreatePollRequest) (*models.Poll, error) {
	if len(req.Options) < 2 || len(req.Options) > 10 {
		return nil, fmt.Errorf("poll must have between 2 and 10 options")
	}
	for i, opt := range req.Options {
		if strings.TrimSpace(opt.Text) == "" {
			return nil, fmt.Errorf("option %d text is required", i+1)
		}
		if len(strings.TrimSpace(opt.Text)) > 120 {
			return nil, fmt.Errorf("option %d text is too long", i+1)
		}
		if strings.TrimSpace(opt.VibeMessage) == "" {
			opt.VibeMessage = defaultVibeMessage(opt.Text)
		}
		if strings.TrimSpace(opt.ID) == "" {
			opt.ID = uuid.NewString()
		}
		req.Options[i].ID = opt.ID
		req.Options[i].VibeMessage = strings.TrimSpace(opt.VibeMessage)
	}

	shareCode, err := generateShareCode()
	if err != nil {
		return nil, err
	}

	poll := &models.Poll{
		ID:                     uuid.NewString(),
		Question:               strings.TrimSpace(req.Question),
		Description:            strings.TrimSpace(req.Description),
		Options:                req.Options,
		CreatedBy:              userID,
		ShareCode:              shareCode,
		IsActive:               true,
		AllowOneVotePerBrowser: req.AllowOneVotePerBrowser,
		AllowAnonymousVoting:   req.AllowAnonymousVoting,
		ShowPercentages:        req.ShowPercentages,
		ShowTotalVoteCount:     req.ShowTotalVoteCount,
		ExpiresAt:              req.ExpiresAt,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *PollService) FindByID(ctx context.Context, id string) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByID(ctx, id)
	if err != nil || poll == nil {
		return poll, err
	}
	return poll, s.ensureOptionIDs(ctx, poll)
}

func (s *PollService) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByShareCode(ctx, shareCode)
	if err != nil || poll == nil {
		return poll, err
	}
	return poll, s.ensureOptionIDs(ctx, poll)
}

func (s *PollService) ListByUser(ctx context.Context, userID string) ([]models.Poll, error) {
	polls, err := s.pollRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range polls {
		if err := s.ensureOptionIDs(ctx, &polls[i]); err != nil {
			return nil, err
		}
	}
	return polls, nil
}

func (s *PollService) ensureOptionIDs(ctx context.Context, poll *models.Poll) error {
	changed := false
	for i := range poll.Options {
		if strings.TrimSpace(poll.Options[i].ID) == "" {
			poll.Options[i].ID = uuid.NewString()
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.pollRepo.Update(ctx, poll.ID, bson.M{"options": poll.Options})
}

func (s *PollService) PollStats(ctx context.Context, pollID, userID string) (map[string]any, error) {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return nil, fmt.Errorf("poll not found")
	}
	if poll.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	total, err := s.voteRepo.CountByPoll(ctx, pollID)
	if err != nil {
		return nil, err
	}
	counts := make([]map[string]any, 0, len(poll.Options))
	for _, option := range poll.Options {
		count, err := s.voteRepo.CountByPollAndOption(ctx, pollID, option.ID)
		if err != nil {
			return nil, err
		}
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) * 100 / float64(total)
		}
		counts = append(counts, map[string]any{
			"optionId":   option.ID,
			"label":      option.Text,
			"count":      count,
			"percentage": percentage,
		})
	}

	return map[string]any{
		"pollId":       pollID,
		"totalVotes":   total,
		"optionCounts": counts,
		"updatedAt":    time.Now(),
	}, nil
}

func (s *PollService) UpdatePoll(ctx context.Context, pollID, userID string, req models.UpdatePollRequest) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return nil, fmt.Errorf("poll not found")
	}
	if poll.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	if req.Question != "" {
		poll.Question = strings.TrimSpace(req.Question)
	}
	if req.Description != "" {
		poll.Description = strings.TrimSpace(req.Description)
	}
	if req.Options != nil {
		if len(req.Options) < 2 || len(req.Options) > 10 {
			return nil, fmt.Errorf("poll must have between 2 and 10 options")
		}
		poll.Options = req.Options
	}
	poll.AllowOneVotePerBrowser = req.AllowOneVotePerBrowser
	poll.AllowAnonymousVoting = req.AllowAnonymousVoting
	poll.ShowPercentages = req.ShowPercentages
	poll.ShowTotalVoteCount = req.ShowTotalVoteCount
	if req.ExpiresAt != nil {
		poll.ExpiresAt = req.ExpiresAt
	}
	poll.UpdatedAt = time.Now()
	if err := s.pollRepo.Update(ctx, pollID, bson.M{
		"question":               poll.Question,
		"description":            poll.Description,
		"options":                poll.Options,
		"allowOneVotePerBrowser": poll.AllowOneVotePerBrowser,
		"allowAnonymousVoting":   poll.AllowAnonymousVoting,
		"showPercentages":        poll.ShowPercentages,
		"showTotalVoteCount":     poll.ShowTotalVoteCount,
		"expiresAt":              poll.ExpiresAt,
	}); err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *PollService) ClosePoll(ctx context.Context, pollID, userID string) error {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return fmt.Errorf("poll not found")
	}
	if poll.CreatedBy != userID {
		return fmt.Errorf("unauthorized")
	}
	poll.IsActive = false
	poll.UpdatedAt = time.Now()
	return s.pollRepo.Update(ctx, pollID, bson.M{"isActive": false})
}

func (s *PollService) DeletePoll(ctx context.Context, pollID, userID string) error {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return fmt.Errorf("poll not found")
	}
	if poll.CreatedBy != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.pollRepo.Delete(ctx, pollID)
}

func (s *PollService) Vote(ctx context.Context, poll *models.Poll, optionID string, voterID string, voterKey string) (models.VoteUpdateEvent, error) {
	if !poll.IsActive {
		return models.VoteUpdateEvent{}, fmt.Errorf("poll is closed")
	}
	if poll.ExpiresAt != nil && !time.Now().Before(*poll.ExpiresAt) {
		return models.VoteUpdateEvent{}, fmt.Errorf("poll has expired")
	}
	if !optionExists(poll.Options, optionID) {
		return models.VoteUpdateEvent{}, fmt.Errorf("invalid option")
	}
	if poll.AllowOneVotePerBrowser && voterKey != "" {
		hasVoted, err := s.voteRepo.HasVoted(ctx, poll.ID, voterKey)
		if err != nil {
			return models.VoteUpdateEvent{}, err
		}
		if hasVoted {
			return models.VoteUpdateEvent{}, fmt.Errorf("duplicate vote")
		}
	}
	if voterID != "" {
		hasVoted, err := s.voteRepo.HasUserVoted(ctx, poll.ID, voterID)
		if err != nil {
			return models.VoteUpdateEvent{}, err
		}
		if hasVoted {
			return models.VoteUpdateEvent{}, fmt.Errorf("duplicate vote")
		}
	}

	vote := &models.Vote{ID: uuid.NewString(), PollID: poll.ID, OptionID: optionID, VoterID: voterID, VoterKey: voterKey, CreatedAt: time.Now()}
	if err := s.voteRepo.Create(ctx, vote); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.VoteUpdateEvent{}, fmt.Errorf("duplicate vote")
		}
		return models.VoteUpdateEvent{}, err
	}
	count, err := s.redis.IncrementVote(ctx, poll.ID, optionID)
	if err != nil {
		return models.VoteUpdateEvent{}, err
	}
	total, err := s.voteRepo.CountByPoll(ctx, poll.ID)
	if err != nil {
		return models.VoteUpdateEvent{}, err
	}
	vibeMessage := ""
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			vibeMessage = opt.VibeMessage
			break
		}
	}
	event := models.VoteUpdateEvent{
		PollID:      poll.ID,
		OptionID:    optionID,
		NewCount:    count,
		TotalVotes:  total,
		Timestamp:   time.Now(),
		VibeMessage: vibeMessage,
	}
	if err := s.redis.PublishVoteUpdate(ctx, event); err != nil {
		return models.VoteUpdateEvent{}, err
	}
	return event, nil
}

func defaultVibeMessage(optionText string) string {
	text := strings.ToLower(strings.TrimSpace(optionText))
	switch {
	case strings.Contains(text, "yes") || strings.Contains(text, "agree") || strings.Contains(text, "go") || strings.Contains(text, "positive"):
		return "✨ Great choice!"
	case strings.Contains(text, "no") || strings.Contains(text, "disagree") || strings.Contains(text, "skip") || strings.Contains(text, "not"):
		return "🌱 No worries! Your choice matters too."
	default:
		return "🤔 Still deciding? That's okay!"
	}
}

func generateShareCode() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func optionExists(options []models.PollOption, target string) bool {
	for _, option := range options {
		if option.ID == target {
			return true
		}
	}
	return false
}
