package common

import (
	"github.com/gin-gonic/gin"
)

// APIError returns an uniform json formatted error
func APIError(c *gin.Context, err string) {
	c.JSON(500, gin.H{
		"status": "error",
		"error":  err,
	})
}
