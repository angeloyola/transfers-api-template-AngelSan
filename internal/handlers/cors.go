package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AllowCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

	// Allow preflight request
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
