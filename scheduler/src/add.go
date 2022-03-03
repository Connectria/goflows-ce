// add.go - add scheduler jobs via the "/api/add" method

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type addType struct {
	TriggerID   string `json:"triggerId"`
	TriggerName string `json:"triggerName"`
	FlowIDs     []struct {
		FlowID          string `json:"flowID"`
		FuncName        string `json:"funcName"`
		FlowDescription string `json:"flowDescription,omitempty"`
		FlowName        string `json:"flowName,omitempty"`
	} `json:"flowIDs"`
	AtList      []int64 `json:"at-list,omitempty"`
	DutyCycle   string  `json:"dutyCycle,omitempty"`
	RepeatUntil int64   `json:"repeat-until,omitempty"`
	Inputs      []struct {
		InputName  string `json:"inputName"`
		InputValue string `json:"inputValue"`
	} `json:"inputs,omitempty"`
}

var schedulerJobRef = make(map[cron.EntryID]addType)
var seq cron.EntryID

// increment job id
func nextInternalJobID() cron.EntryID {
	seq++
	return seq
}

// decrement job id (on error)
func reduceInternalJobID() {
	seq--
}

// AtSchedule used for at jobs
type atSchedule struct {
	Delay time.Duration
}

// At returns the durations which should be the same
func (sched atSchedule) At(t time.Time) atSchedule {
	return atSchedule{
		Delay: time.Until(t),
	}
}

// Next is implemented
func (sched atSchedule) Next(t time.Time) time.Time {
	return t.Add(sched.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
}

// handleAdd
func handleAdd(c *gin.Context) {
	var logMsg string
	var add addType

	if err := c.ShouldBindJSON(&add); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// can only be cron or at -- not both
	if (len(add.DutyCycle) > 0) && (len(add.AtList) > 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Use 'at-list' or 'dutyCycle' - cannot specifiy both"})
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

	// schedule the "at" job
	if len(add.AtList) > 0 {
		var schedulerJobIDList []cron.EntryID

		// add the list of At Jobs to the scheduler
		for _, at := range add.AtList {

			// is this in the past?
			now := time.Now().UTC()
			if at < now.Unix() {
				logMsg := fmt.Sprintf("Job will not be scheduled due to atTime ('%v') less than current time ('%v')", time.Unix(at, 0).UTC(), time.Unix(now.Unix(), 0).UTC())
				logger.Info().
					Str("function", "func()").
					Str("triggerId", add.TriggerID).
					Msg(logMsg)
				c.JSON(http.StatusBadRequest, gin.H{"error": logMsg})
				return
			}

			// add new job
			schedulerJobID := nextInternalJobID()
			s := new(atSchedule)
			internalJobID := scheduler.Schedule(s.At(time.Unix(at, 0).UTC()), cron.FuncJob(func() {

				sendTask, err := json.Marshal(add)
				if err != nil {
					logger.Error().
						Str("function", "func()").
						Msgf("json.Marshal() returned '%s' when trying to marshal:  %v", err.Error(), add)
					return
				}

				// past our time?
				now := time.Now().UTC()
				if at < now.Unix() {
					logger.Info().
						Str("function", "func()").
						Str("triggerId", add.TriggerID).
						Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
						Msgf("Current time ('%v') exceeds at('%v'); removing from scheduler", time.Unix(now.Unix(), 0).UTC(), time.Unix(at, 0).UTC())
					scheduler.Remove(schedulerJobID)
					return
				}

				// send to goflows-processor
				err = publishRabbitMQ(sendTask)
				if err != nil {
					logger.Error().
						Str("function", "func()").
						Str("triggerId", add.TriggerID).
						Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
						Msgf("publishRabbitMQ() returned '%s' when trying to send '%v' to goflows-processor.", err.Error(), add.TriggerName)
					return
				}

				logger.Info().
					Str("function", "func()").
					Str("triggerId", add.TriggerID).
					Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
					Msgf("Sent '%v' to goflows-processor for execution.", add.TriggerID)
			}))

			// make sure IDs match
			if internalJobID != schedulerJobID {
				logger.Error().
					Str("function", "handleAdd()").
					Str("triggerId", add.TriggerID).
					Str("internalJobID", fmt.Sprintf("%v", internalJobID)).
					Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
					Msgf("internalJobID and schedulerJobID do not match; removed '%v' from scheduler", add.TriggerName)

				c.JSON(http.StatusInternalServerError,
					gin.H{"status": "internalJobID and schedulerJobID do not match",
						"atTime":         at,
						"flowIDs":        add.FlowIDs,
						"internalJobID":  internalJobID,
						"schedulerJobID": schedulerJobID,
						"triggerId":      add.TriggerID,
					})
				scheduler.Remove(internalJobID)
				return
			}

			// update cross-reference
			schedulerJobRef[internalJobID] = add
			schedulerJobIDList = append(schedulerJobIDList, schedulerJobID)
			logger.Info().
				Str("function", "handleAdd()").
				Str("triggerId", add.TriggerID).
				Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
				Msgf("Added '%v' at-job to scheduler", add.TriggerName)
		}

		// return a list of At Jobs that were added
		c.JSON(http.StatusOK,
			gin.H{"status": "added jobs to scheduler",
				"atList":             add.AtList,
				"flowIDs":            add.FlowIDs,
				"schedulerJobIDList": schedulerJobIDList,
				"triggerId":          add.TriggerID,
			})
		return
	}

	// schedule the "cron" job
	if len(add.DutyCycle) > 0 {
		schedulerJobID := nextInternalJobID()
		internalJobID, err := scheduler.AddFunc(add.DutyCycle, func() {

			sendTask, err := json.Marshal(add)
			if err != nil {
				logger.Error().
					Str("function", "func()").
					Msgf("json.Marshal() returned '%s' when trying to marshal:  %v", err.Error(), add)
				return
			}

			// past our time?
			now := time.Now().UTC()
			if add.RepeatUntil < now.Unix() {
				logger.Info().
					Str("function", "func()").
					Str("triggerId", add.TriggerID).
					Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
					Msgf("Current time ('%v') exceeds repeat-until ('%v'); removing from scheduler", time.Unix(now.Unix(), 0).UTC(), time.Unix(add.RepeatUntil, 0).UTC())
				scheduler.Remove(schedulerJobID)
				return
			}

			// send to goflows-processor
			err = publishRabbitMQ(sendTask)
			if err != nil {
				logger.Error().
					Str("function", "func()").
					Str("triggerId", add.TriggerID).
					Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
					Msgf("publishRabbitMQ() returned '%s' when trying to send '%v' to goflows-processor.", err.Error(), add.TriggerName)
				return
			}

			logger.Info().
				Str("function", "func()").
				Str("triggerId", add.TriggerID).
				Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
				Msgf("Sent '%v' to goflows-processor for execution.", add.TriggerName)
		})

		if err != nil {
			reduceInternalJobID()
			logMsg = fmt.Sprintf("scheduler.AddFunc() returned '%s'", err.Error())
			logger.Error().
				Str("function", "handleAdd()").
				Msg(logMsg)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": logMsg})
			return
		}

		// make sure IDs match
		if internalJobID != schedulerJobID {
			logger.Error().
				Str("function", "handleAdd()").
				Str("triggerId", add.TriggerID).
				Str("internalJobID", fmt.Sprintf("%v", internalJobID)).
				Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
				Msgf("internalJobID and schedulerJobID do not match; removed '%v' from scheduler", add.TriggerName)

			c.JSON(http.StatusInternalServerError,
				gin.H{"status": "internalJobID and schedulerJobID do not match",
					"dutyCycle":      add.DutyCycle,
					"flowIDs":        add.FlowIDs,
					"internalJobID":  internalJobID,
					"schedulerJobID": schedulerJobID,
					"triggerId":      add.TriggerID,
				})
			scheduler.Remove(internalJobID)
			return
		}

		// update
		schedulerJobRef[internalJobID] = add
		logger.Info().
			Str("function", "handleAdd()").
			Str("triggerId", add.TriggerID).
			Str("schedulerJobID", fmt.Sprintf("%v", schedulerJobID)).
			Msgf("Added '%v' at-job to scheduler", add.TriggerName)

		c.JSON(http.StatusOK,
			gin.H{"status": "added to scheduler",
				"dutyCycle":      add.DutyCycle,
				"repeat-until":   add.RepeatUntil,
				"flowIDs":        add.FlowIDs,
				"schedulerJobID": schedulerJobID,
				"triggerId":      add.TriggerID,
			})
		return
	}

	// atTime or dutyCycle not specified
	c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required field: 'at-list' or 'dutyCycle' - one must be specified"})
}
