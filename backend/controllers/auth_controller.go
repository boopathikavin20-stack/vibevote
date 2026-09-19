package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pulsevote/models"
	"pulsevote/services"
	"pulsevote/utils"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) Signup(ctx *gin.Context) {
	var req models.SignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, "Invalid signup payload")
		return
	}
	user, token, err := c.authService.Signup(context.Background(), req)
	if err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusCreated, "Account created successfully", gin.H{"user": user, "token": token, "expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339)})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.JSONError(ctx, http.StatusBadRequest, "Invalid login payload")
		return
	}
	user, token, err := c.authService.Login(context.Background(), req)
	if err != nil {
		utils.JSONError(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "Login successful", gin.H{"user": user, "token": token, "expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339)})
}

func (c *AuthController) Me(ctx *gin.Context) {
	userID, exists := ctx.Get("userId")
	if !exists {
		utils.JSONError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, err := c.authService.Me(context.Background(), userID.(string))
	if err != nil {
		utils.JSONError(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	utils.JSONSuccess(ctx, http.StatusOK, "User profile loaded", gin.H{"user": user})
}
