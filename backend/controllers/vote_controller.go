package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"pulsevote/models"
	"pulsevote/services"
	"pulsevote/utils"
)

type VoteController struct {
	pollService *services.PollService
}

func NewVoteController(pollService *services.PollService) *VoteController {
	return &VoteController{pollService: pollService}
}

func (c *VoteController) Vote(ctx *gin.Context) {
	pollID := ctx.Param("id")
	poll, err := c.pollService.FindByID(context.Background(), pollID)
	if err != nil || poll == nil {
		utils.JSONError(ctx, http.StatusNotFound, "Poll not found")
		return
	}
	var req models.VoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, "Invalid vote payload")
		return
	}
	voterID := ""
	if userID, exists := ctx.Get("userId"); exists {
		if id, ok := userID.(string); ok {
			voterID = id
		}
	}
	voterKey := req.VoterKey
	if voterID != "" {
		voterKey = ""
	} else if poll.AllowOneVotePerBrowser && strings.TrimSpace(voterKey) == "" {
		utils.JSONError(ctx, http.StatusBadRequest, "A voter identity is required")
		return
	}
	result, err := c.pollService.Vote(context.Background(), poll, req.OptionID, voterID, voterKey)
	if err != nil {
		if strings.EqualFold(err.Error(), "duplicate vote") {
			utils.JSONError(ctx, http.StatusConflict, "You have already voted in this poll.")
			return
		}
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusCreated, "Vote recorded successfully", gin.H{"vote": result, "vibeMessage": result.VibeMessage, "timestamp": time.Now().Format(time.RFC3339)})
}
