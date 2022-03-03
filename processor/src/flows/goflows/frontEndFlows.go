// frontEndFlows.go	- functions for flows

package goflows

import (
	"encoding/json"
)

// declare trigger functions for GoFlows
var FrontEndFlows = map[string]interface{}{
	"GetFlow": GetFlow,
}

// FlowType is received from the front end API
type FlowType struct {
	CreatedOn       string `json:"createdOn"`
	FlowDescription string `json:"flowDescription"`
	FlowDocument    struct {
		FuncName string `json:"funcName"`
	} `json:"flowDocument"`
	FlowID               string          `json:"flowId"`
	FlowInputs           []FlowInputType `json:"flowInputs"`
	FlowName             string          `json:"flowName"`
	FlowOutputs          []interface{}   `json:"flowOutputs"`
	FlowShortDescription string          `json:"flowShortDescription"`
	UpdatedOn            string          `json:"updatedOn"`
}

// GetFlow retrieves a flow using the front end API
func GetFlow(flowId string) (FlowType, error) {
	var flow FlowType

	body, err := frontEndAPI("GET", "/api/flows/"+flowId, nil)
	if err != nil {
		return FlowType{}, err
	}

	json.Unmarshal(body, &flow)
	return flow, nil
}
