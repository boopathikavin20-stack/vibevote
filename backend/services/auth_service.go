package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"pulsevote/middleware"
	"pulsevote/models"
	"pulsevote/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Signup(ctx context.Context, req models.SignupRequest) (*models.User, string, error) {
	if strings.TrimSpace(req.Name) == "" || len(req.Name) < 2 {
		return nil, "", fmt.Errorf("name is required")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		return nil, "", fmt.Errorf("email is required")
	}
	if len(req.Password) < 6 {
		return nil, "", fmt.Errorf("password must be at least 6 characters")
	}
	if existing, err := s.userRepo.FindByEmail(ctx, req.Email); err != nil {
		return nil, "", err
	} else if existing != nil {
		return nil, "", fmt.Errorf("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{ID: req.Email, Name: req.Name, Email: req.Email, Password: string(hash), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}
	jwtToken, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}
	user.Password = ""
	return user, jwtToken, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.User, string, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}
	jwtToken, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}
	user.Password = ""
	return user, jwtToken, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	user.Password = ""
	return user, nil
}
