// validate.go - validate cron-expression

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type validateType struct {
	DutyCycle string `json:"dutyCycle"`
}

// handleValidate
func handleValidate(c *gin.Context) {
	var presented validateType

	if err := c.ShouldBindJSON(&presented); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// dutyCycle not specified
	if presented.DutyCycle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required field: 'dutyCycle'"})
		return
	}

	// test the cron-expression
	t := cron.New()
	_, err := t.AddFunc(presented.DutyCycle, func() {})
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{
				"dutyCycle": presented.DutyCycle,
				"error":     err.Error(),
				"status":    "invalid",
			})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"status":    "valid",
			"dutyCycle": presented.DutyCycle,
		})
}
