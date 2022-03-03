// remove.go - remove a scheduled job via the "/api/remove" method

package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// handleRemove
func handleRemove(c *gin.Context) {
	var removeScheduledJobID cron.EntryID

	// for "/api/remove/schedudulerJobID/:schedulerJobID"
	if len(c.Params.ByName("schedulerJobID")) > 0 {
		schedulerJobID, err := strconv.ParseInt(c.Params.ByName("schedulerJobID"), 10, 64)
		if err != nil {
			logger.Warn().
				Str("function", "handleRemove()").
				Msgf("strconv.ParseInt() returned '%s' when converting c.Params.ByName()", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if schedulerJobID > 0 {
			removeScheduledJobID = cron.EntryID(schedulerJobID)
			_, ok := schedulerJobRef[removeScheduledJobID]
			if !ok {
				logger.Info().
					Str("function", "handleRemove()").
					Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
					Msg("requested schedulerJobID not found")
				c.JSON(http.StatusOK,
					gin.H{"error": "schedulerJobID not found",
						"schedulerJobID": removeScheduledJobID,
					})
				return
			}

			// remove the job
			scheduler.Remove(removeScheduledJobID)
			delete(schedulerJobRef, removeScheduledJobID)
			logger.Info().
				Str("function", "handleRemove()").
				Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
				Msg("removed from goflows-scheduler")
			c.JSON(http.StatusOK,
				gin.H{"status": "removed from goflows-scheduler",
					"schedulerJobID": removeScheduledJobID,
				})
			return
		}
	}

	// for "/api/remove/triggerId/:triggerId"
	triggerId := c.Params.ByName("triggerId")
	if len(triggerId) > 0 {
		for _, k := range scheduler.Entries() {
			if triggerId == schedulerJobRef[k.ID].TriggerID {
				removeScheduledJobID = k.ID
				break
			}
		}

		_, ok := schedulerJobRef[removeScheduledJobID]
		if !ok {
			logger.Info().
				Str("function", "handleRemove()").
				Str("triggerId", triggerId).
				Msg("requested triggerId not found")
			c.JSON(http.StatusOK,
				gin.H{"error": "triggerId not found",
					"triggerId": triggerId,
				})
			return
		}

		// remove the job
		scheduler.Remove(removeScheduledJobID)
		delete(schedulerJobRef, removeScheduledJobID)
		logger.Info().
			Str("function", "handleRemove()").
			Str("triggerId", triggerId).
			Msg("removed from goflows-scheduler")
		c.JSON(http.StatusOK,
			gin.H{"status": "removed from goflows-scheduler",
				"triggerId": triggerId,
			})
		return
	}

}
