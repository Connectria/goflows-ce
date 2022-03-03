// validateRules.go	- test an event against event rules

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleValidateRulesPut
func handleApiValidateRulesGet(c *gin.Context) {
	if len(c.Query("eventAlertID")) > 0 {

		thisEvent, err := lookupEventAlertID(c.Query("eventAlertID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        err.Error(),
				"eventAlertID": c.Query("eventAlertID"),
				"status":       "not-matched",
			})
			return
		}

		// unknown event action
		if len(getOpsGenieEventFieldValue(thisEvent, "SQSMessageID")) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        err.Error(),
				"eventAlertID": c.Query("eventAlertID"),
				"status":       "empty OpsGenie event",
			})
			return
		}

		msgID := getOpsGenieEventFieldValue(thisEvent, "SQSMessageID")
		evalResp, err := evalRules(msgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     err.Error(),
				"messageID": msgID,
				"status":    "not-matched",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"eventAlertID": evalResp.eventAlertID,
			"messageID":    msgID,
			"setVars":      evalResp.setVars,
			"status":       "matched",
			"triggerID":    evalResp.trigger.TriggerID,
			"triggerName":  evalResp.trigger.Name,
		})
		return
	}

	if len(c.Query("messageID")) > 0 {

		evalResp, err := evalRules(c.Query("messageID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     err.Error(),
				"messageID": c.Query("messageID"),
				"status":    "not-matched",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"eventAlertID": evalResp.eventAlertID,
			"messageID":    c.Query("messageID"),
			"setVars":      evalResp.setVars,
			"status":       "matched",
			"triggerID":    evalResp.trigger.TriggerID,
			"triggerName":  evalResp.trigger.Name,
		})
		return
	}

	// ought oh
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "messageID not specified",
	})
}
