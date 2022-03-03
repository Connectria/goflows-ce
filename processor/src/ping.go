// ping.go

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handlePing
func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "pong"})
}
