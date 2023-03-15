package controllers

import "github.com/gin-gonic/gin"

func ResponseSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"data":    data,
		"message": message,
	})
}

func ResponseError(c *gin.Context, code int, message string, error interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"error":   error,
		"message": message,
	})
}
