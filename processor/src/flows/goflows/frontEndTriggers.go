// frontEndTriggers.go	- functions for triggers

package goflows

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// declare trigger functions for GoFlows
var FrontEndTriggers = map[string]interface{}{
	"CreateTrigger":  CreateTrigger,
	"GetTrigger":     GetTrigger,
	"UpdateTrigger":  UpdateTrigger,
	"DeleteTrigger":  DeleteTrigger,
	"GetAllTriggers": GetAllTriggers,
}

// TriggerType is used to interact with front end API
type TriggerType struct {
	Name         string          `json:"name"`
	Active       bool            `json:"active"`
	Flowids      []string        `json:"flowIds"`
	TriggerID    string          `json:"triggerId"`
	Triggerids   []string        `json:"triggerIds"`
	Inputs       []FlowInputType `json:"inputs"`
	Triggerlogic struct {
		EventRules     []json.RawMessage `json:"event-rules,omitempty"`
		Triggersubtype string            `json:"triggerSubType,omitempty"`
		Triggertype    string            `json:"triggerType,omitempty"`
		TemporalRule   struct {
			AtList         []int `json:"at-list,omitempty"`
			CronExpression struct {
				Expression  string `json:"expression,omitempty"`
				RepeatUntil int    `json:"repeat-until,omitempty"`
			} `json:"cron-expression,omitempty"`
		} `json:"temporal-rule"`
	} `json:"triggerLogic"`
	CreatedOn string `json:"createdOn,omitempty"`
	UpdatedOn string `json:"updatedOn,omitempty"`
	Weight    int64
}

// TriggerSetup is used to init the TriggerType for creates
type TriggerSetup struct {
	Name           string
	FlowID         string
	Inputs         []FlowInputType
	CronExpression string
	RepeatUntil    string
	TriggerID      string
}

// TriggerInit is used to init the trigger struct for creation/update triggers by triggers
func TriggerTriggerSetup(ts TriggerSetup) TriggerType {

	t := TriggerType{}
	t.Active = true
	t.Name = ts.Name
	t.Flowids = []string{ts.FlowID}
	t.Inputs = removeReservedFromJobInputVars(ts.Inputs)
	t.TriggerID = ts.TriggerID
	t.Triggerids = []string{}
	t.Triggerlogic.TemporalRule.AtList = []int{}
	t.Triggerlogic.TemporalRule.CronExpression.Expression = ts.CronExpression

	endDate, _ := strconv.Atoi(ts.RepeatUntil)
	if endDate > 0 {
		t.Triggerlogic.TemporalRule.CronExpression.RepeatUntil = endDate
	}

	return t
}

// CreateTrigger creates a new trigger using the front end API
func CreateTrigger(t TriggerType) (TriggerType, error) {
	var trigger TriggerType

	jsonData, _ := json.Marshal(t)
	body, err := frontEndAPI("POST", "/api/triggers", jsonData)
	if err != nil {
		return TriggerType{}, err
	}

	json.Unmarshal(body, &trigger)
	return trigger, nil
}

// UpdateTrigger updates a trigger using the front end API
func UpdateTrigger(t TriggerType) (TriggerType, error) {

	if len(t.TriggerID) == 0 {
		return TriggerType{}, fmt.Errorf("t.TriggerID is empty")
	}

	var trigger TriggerType

	jsonData, _ := json.Marshal(t)
	body, err := frontEndAPI("PATCH", "/api/triggers/"+t.TriggerID, jsonData)
	if err != nil {
		return TriggerType{}, err
	}

	json.Unmarshal(body, &trigger)
	return trigger, nil
}

// DeleteTrigger removes a trigger using the front end API
func DeleteTrigger(triggerId string) error {
	_, err := frontEndAPI("DELETE", "/api/triggers/"+triggerId, nil)
	return err
}

// GetTrigger retrieves a trigger using the front end API
func GetTrigger(triggerId string) (TriggerType, error) {
	var trigger TriggerType

	body, err := frontEndAPI("GET", "/api/triggers/"+triggerId, nil)
	if err != nil {
		return TriggerType{}, err
	}

	json.Unmarshal(body, &trigger)
	return trigger, nil
}

// GetAllTriggers retrieves all the triggers using the front end API
func GetAllTriggers() (interface{}, error) {
	var allTriggers []TriggerType

	body, err := frontEndAPI("GET", "/api/triggers", nil)
	if err != nil {
		return "", err
	}

	json.Unmarshal(body, &allTriggers)
	return allTriggers, nil
}
