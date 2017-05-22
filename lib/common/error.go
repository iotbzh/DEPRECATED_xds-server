package common

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// APIError returns an uniform json formatted error
func APIError(c *gin.Context, format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	c.JSON(500, gin.H{
		"status": "error",
		"error":  errMsg,
	})
}
