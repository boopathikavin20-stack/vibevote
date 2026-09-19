package utils

import "github.com/gin-gonic/gin"

func JSONSuccess(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func JSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"data":    nil,
	})
}
