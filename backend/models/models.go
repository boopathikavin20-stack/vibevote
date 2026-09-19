package models

import (
	"time"
)

type User struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"passwordHash"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type PollOption struct {
	ID          string `json:"id" bson:"id"`
	Text        string `json:"text" bson:"text"`
	VibeMessage string `json:"vibeMessage" bson:"vibeMessage"`
}

type Poll struct {
	ID                     string       `json:"id" bson:"_id,omitempty"`
	Question               string       `json:"question" bson:"question"`
	Description            string       `json:"description,omitempty" bson:"description,omitempty"`
	Options                []PollOption `json:"options" bson:"options"`
	CreatedBy              string       `json:"createdBy" bson:"createdBy"`
	ShareCode              string       `json:"shareCode" bson:"shareCode"`
	IsActive               bool         `json:"isActive" bson:"isActive"`
	AllowOneVotePerBrowser bool         `json:"allowOneVotePerBrowser" bson:"allowOneVotePerBrowser"`
	AllowAnonymousVoting   bool         `json:"allowAnonymousVoting" bson:"allowAnonymousVoting"`
	ShowPercentages        bool         `json:"showPercentages" bson:"showPercentages"`
	ShowTotalVoteCount     bool         `json:"showTotalVoteCount" bson:"showTotalVoteCount"`
	ExpiresAt              *time.Time   `json:"expiresAt,omitempty" bson:"expiresAt,omitempty"`
	CreatedAt              time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt              time.Time    `json:"updatedAt" bson:"updatedAt"`
}

type Vote struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	PollID    string    `json:"pollId" bson:"pollId"`
	OptionID  string    `json:"optionId" bson:"optionId"`
	VoterID   string    `json:"voterId,omitempty" bson:"voterId,omitempty"`
	VoterKey  string    `json:"voterKey,omitempty" bson:"voterKey,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type CreatePollRequest struct {
	Question               string       `json:"question" binding:"required,min=6,max=200"`
	Description            string       `json:"description,omitempty"`
	Options                []PollOption `json:"options" binding:"required,dive,required"`
	AllowOneVotePerBrowser bool         `json:"allowOneVotePerBrowser"`
	AllowAnonymousVoting   bool         `json:"allowAnonymousVoting"`
	ShowPercentages        bool         `json:"showPercentages"`
	ShowTotalVoteCount     bool         `json:"showTotalVoteCount"`
	ExpiresAt              *time.Time   `json:"expiresAt,omitempty"`
}

type UpdatePollRequest struct {
	Question               string       `json:"question,omitempty"`
	Description            string       `json:"description,omitempty"`
	Options                []PollOption `json:"options,omitempty"`
	AllowOneVotePerBrowser bool         `json:"allowOneVotePerBrowser"`
	AllowAnonymousVoting   bool         `json:"allowAnonymousVoting"`
	ShowPercentages        bool         `json:"showPercentages"`
	ShowTotalVoteCount     bool         `json:"showTotalVoteCount"`
	ExpiresAt              *time.Time   `json:"expiresAt,omitempty"`
}

type VoteRequest struct {
	OptionID string `json:"optionId" binding:"required"`
	VoterKey string `json:"voterKey,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=80"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type TokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
}

type VoteUpdateEvent struct {
	PollID      string    `json:"pollId"`
	OptionID    string    `json:"optionId"`
	NewCount    int64     `json:"newCount"`
	TotalVotes  int64     `json:"totalVotes"`
	Timestamp   time.Time `json:"timestamp"`
	VibeMessage string    `json:"vibeMessage,omitempty"`
}
