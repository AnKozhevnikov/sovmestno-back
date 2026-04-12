package handlers

import "github.com/gin-gonic/gin"

func errResponse(code, message string) gin.H {
	return gin.H{
		"errors": []gin.H{
			{"code": code, "message": message},
		},
	}
}
