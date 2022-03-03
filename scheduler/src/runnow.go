// runnow.go - immediate execution of the specified funcname

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleRunNow godoc
// @Summary Immediately schedule GoFlow
// @Description Send the provided information immediately to the goflows-processor for execution. Must include triggerId, flowID and funcName; cron-expression, repeat-util, and at-list are ignored.
// @Accept json
// @Param block body addType true "must include triggerId and flowID"
// @Produce json
// @Success 200
// @Failure 400
// @Router /api/runNow [post]
func handleRunNow(c *gin.Context) {
	var logMsg string
	var add addType

	if err := c.ShouldBindJSON(&add); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// must include triggerId
	if len(add.TriggerID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required field: 'triggerId'"})
		return
	}

	// must include funcName
	if len(add.FlowIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required list of 'flowIDs'"})
		return
	}

	// set up similar to schedulded jobs
	sendTask, err := json.Marshal(add)
	if err != nil {
		logger.Error().
			Str("function", "func()").
			Msgf("json.Marshal() returned '%s' when trying to marshal:  %v", err.Error(), add)
		return
	}

	// send to goflows-processor
	err = publishRabbitMQ(sendTask)
	if err != nil {
		logMsg = fmt.Sprintf("publishRabbitMQ() returned '%s' when trying to send '%v' to goflows-processor.", err.Error(), add.TriggerName)
		logger.Error().
			Str("function", "func()").
			Str("triggerId", add.TriggerID).
			Msg(logMsg)

		c.JSON(http.StatusInternalServerError,
			gin.H{"status": logMsg,
				"flowIDs":   add.FlowIDs,
				"triggerId": add.TriggerID,
			})
		return
	}

	logger.Info().
		Str("function", "func()").
		Str("triggerId", add.TriggerID).
		Msgf("Sent '%v' to goflows-processor for execution.", add.TriggerID)

	c.JSON(http.StatusOK,
		gin.H{"status": "Sent to goflows-processor for execution",
			"flowIDs":   add.FlowIDs,
			"triggerId": add.TriggerID,
		})
}
