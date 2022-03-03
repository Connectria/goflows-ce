// setTriggerEventTTL.go	- change the Trigger Event TTL used by the cache without restarting the dameon

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// handleSetTriggerEventTTL
func handleApiSetTriggerEventTTL(c *gin.Context) {
	if len(c.Query("TTL")) > 0 {
		ttl, err := strconv.ParseInt(c.Query("TTL"), 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "invalid TTL.",
			})
			return
		}

		// let's not press our luck
		if ttl < 30 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "TTL must be greater than 30",
			})
			return
		}

		// update the TTL and refresh
		triggerEventRules.TTL = ttl
		err = triggerEventRules.UpdateEventTriggersCache()
		if err != nil {
			log.Fatalf("%v", err.Error())
		}

		c.JSON(http.StatusOK, gin.H{
			"TTL":     triggerEventRules.TTL,
			"message": "event triggers updated",
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf("'%v' unknown", c.Query("TTL")),
	})
}
