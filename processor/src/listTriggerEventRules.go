// listTriggerEventRules.go	- return the TriggerEventRulesCache

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleApiListTriggerEventRules
func handleApiListTriggerEventRules(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"count":      len(triggerEventRules.Triggers),
			"triggers":   triggerEventRules.Triggers,
			"lastUpdate": triggerEventRules.LastUpdate,
			"TTL":        triggerEventRules.TTL,
		})
}
