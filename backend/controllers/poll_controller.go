package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pulsevote/middleware"
	"pulsevote/models"
	"pulsevote/services"
	"pulsevote/utils"
)

type PollController struct {
	pollService *services.PollService
}

func NewPollController(pollService *services.PollService) *PollController {
	return &PollController{pollService: pollService}
}

func (c *PollController) Create(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.CreatePollRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, "Invalid poll payload")
		return
	}
	if strings.TrimSpace(req.Question) == "" {
		utils.JSONError(ctx, http.StatusBadRequest, "Question is required")
		return
	}
	poll, err := c.pollService.CreatePoll(context.Background(), userID, req)
	if err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusCreated, "Poll created successfully", gin.H{"poll": poll})
}

func (c *PollController) List(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	polls, err := c.pollService.ListByUser(context.Background(), userID)
	if err != nil {
		utils.JSONError(ctx, http.StatusInternalServerError, "Unable to list polls")
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Polls loaded", gin.H{"polls": polls})
}

func (c *PollController) Stats(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	stats, err := c.pollService.PollStats(context.Background(), ctx.Param("id"), userID)
	if err != nil {
		utils.JSONError(ctx, http.StatusForbidden, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Poll stats loaded", stats)
}

func (c *PollController) GetPublic(ctx *gin.Context) {
	shareCode := ctx.Param("shareCode")
	poll, err := c.pollService.FindByShareCode(context.Background(), shareCode)
	if err != nil || poll == nil {
		utils.JSONError(ctx, http.StatusNotFound, "Poll not found")
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Poll loaded", gin.H{"poll": poll})
}

func (c *PollController) Update(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	pollID := ctx.Param("id")
	var req models.UpdatePollRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, "Invalid poll update payload")
		return
	}
	poll, err := c.pollService.UpdatePoll(context.Background(), pollID, userID, req)
	if err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Poll updated successfully", gin.H{"poll": poll})
}

func (c *PollController) Delete(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	pollID := ctx.Param("id")
	if err := c.pollService.DeletePoll(context.Background(), pollID, userID); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Poll deleted successfully", gin.H{})
}

func (c *PollController) Close(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == "" {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	pollID := ctx.Param("id")
	if err := c.pollService.ClosePoll(context.Background(), pollID, userID); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Poll closed successfully", gin.H{})
}
