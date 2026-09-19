package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"

	"pulsevote/utils"
)

var jwtSecret = []byte("pulsevote-dev-secret")

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.JSONError(c, 401, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.JSONError(c, 401, "Invalid token format")
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(parts[1], jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			utils.JSONError(c, 401, "Invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.JSONError(c, 401, "Invalid token payload")
			c.Abort()
			return
		}

		userID, userIDOK := claims["userId"].(string)
		userEmail, userEmailOK := claims["email"].(string)
		if !userIDOK || !userEmailOK {
			utils.JSONError(c, 401, "Invalid token payload")
			c.Abort()
			return
		}
		c.Set("userId", userID)
		c.Set("userEmail", userEmail)
		c.Next()
	}
}

func GenerateToken(userID string, email string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	})
	return jwtToken.SignedString(jwtSecret)
}

func GetCurrentUserID(c *gin.Context) string {
	if userID, ok := c.Get("userId"); ok {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
