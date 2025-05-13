package utils

import "github.com/gin-gonic/gin"

func Error(c *gin.Context, httpCode int, err error) {
	c.JSON(httpCode, gin.H{
		"success":    false,
		"statusCode": httpCode,
		"message":    err.Error(),
	})
}

func SuccessWithData(c *gin.Context, httpCode int, data interface{}, message string) {
	c.JSON(httpCode, gin.H{
		"success":    true,
		"statusCode": httpCode,
		"message":    message,
		"data":       data,
	})
}

func SuccessWithoutData(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, gin.H{
		"success":    true,
		"statusCode": httpCode,
		"message":    message,
	})
}
