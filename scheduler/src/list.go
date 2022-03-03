// list.go - list current scheduled jobs via the "/api/list" method

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type listType struct {
	SchedulerJobID int    `json:"schedulerJobID"`
	TriggerID      string `json:"triggerId"`
	TriggerName    string `json:"triggerName"`
	FlowIDs        []struct {
		FlowID          string `json:"flowID"`
		FuncName        string `json:"funcName"`
		FlowDescription string `json:"flowDescription,omitempty"`
		FlowName        string `json:"flowName,omitempty"`
	} `json:"flowIDs"`
	Inputs []struct {
		InputName  string `json:"inputName"`
		InputValue string `json:"inputValue"`
	} `json:"inputs,omitempty"`
	AtList           []int64 `json:"at-list,omitempty"`
	DutyCycle        string  `json:"dutyCycle,omitempty"`
	RepeatUntil      int64   `json:"repeat-until,omitempty"`
	RepeatUntilHuman string  `json:"repeat-untilHuman,omitempty"`
	Next             int64   `json:"next"`
	NextHuman        string  `json:"nextHuman"`
	Prev             int64   `json:"prev"`
	PrevHuman        string  `json:"prevHuman"`
}

// handleList
func handleList(c *gin.Context) {
	var listEntries []*listType
	for _, k := range scheduler.Entries() {
		var listEntry listType

		// list only matching triggerId
		if len(c.Query("triggerId")) > 0 {
			if c.Query("triggerId") != schedulerJobRef[k.ID].TriggerID {
				continue
			}
		}

		// list only match schedulerJobID
		if len(c.Query("schedulerJobID")) > 0 {
			if c.Query("schedulerJobID") != fmt.Sprintf("%v", k.ID) {
				continue
			}
		}

		listEntry.SchedulerJobID = int(k.ID)

		listEntry.TriggerID = schedulerJobRef[k.ID].TriggerID
		listEntry.TriggerName = schedulerJobRef[k.ID].TriggerName

		listEntry.FlowIDs = schedulerJobRef[k.ID].FlowIDs

		listEntry.DutyCycle = schedulerJobRef[k.ID].DutyCycle
		listEntry.RepeatUntil = schedulerJobRef[k.ID].RepeatUntil
		if listEntry.RepeatUntil > 0 {
			listEntry.RepeatUntilHuman = fmt.Sprintf("%v", time.Unix(schedulerJobRef[k.ID].RepeatUntil, 0).UTC())
		}

		listEntry.Inputs = schedulerJobRef[k.ID].Inputs

		listEntry.Next = k.Next.Unix()
		listEntry.NextHuman = fmt.Sprintf("%v", k.Next)

		listEntry.Prev = k.Prev.Unix()
		listEntry.PrevHuman = fmt.Sprintf("%v", k.Prev)

		listEntries = append(listEntries, &listEntry)
	}

	c.JSON(http.StatusOK, gin.H{"scheduler-list": listEntries})
}
