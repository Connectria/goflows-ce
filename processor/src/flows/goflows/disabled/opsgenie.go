// opsgenie.go	- functions for GoFlows to peform OpsGenie actions

package goflows

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

// declare OpsGenie functions for GoFlows
var OpsgenieFuncs = map[string]interface{}{
	"Acknowledge":      Acknowledge,
	"SetExtraProperty": SetExtraProperty,
	"AddTags":          AddTags,
	"Assign":           Assign,
	"AddNote":          AddNote,
}

// Acknowledge will acknowledge an OpsGenie alert
func Acknowledge(alertid string) (string, error) {
	results, err := opsgenieAPI(alertid, "acknowledge", "{\"source\":\"opsgenie-processor\",\"note\":\"GoFlows is acknowledging this OpsGenie alert.\"}")
	return fmt.Sprintf("OpsGenie ACTION CALLED: Acknowledge('%v')", results), err
}

// SetExtraProperty will update an OpsGenie alert extra property Key:Value pair
func SetExtraProperty(alertid, key, value string) (string, error) {
	data := fmt.Sprintf("{\"source\":\"opsgenie-processor\",\"details\":{\"%v\":\"%v\"}}", key, value)
	results, err := opsgenieAPI(alertid, "details", data)
	return fmt.Sprintf("OpsGenie ACTION CALLED: SetExtraProperty('%v', '%v') %v", key, value, results), err
}

// AddTags will update an OpsGenie alert with a tag
func AddTags(alertid string, tags ...string) (string, error) {
	data := fmt.Sprintf("{\"source\":\"opsgenie-processor\",\"tags\":%v}", buildArray(tags))
	results, err := opsgenieAPI(alertid, "tags", data)
	return fmt.Sprintf("OpsGenie ACTION CALLED: AddTags('%v') %v", tags, results), err
}

// Assign will assign an OpsGenie alert to a team or person
func Assign(alertid, team string) (string, error) {
	return fmt.Sprintf("OpsGenie ACTION CALLED: Assign('%v') - This function has not been implemented.", team), nil
}

// AddNote will add a note to an OpsGenie alert
func AddNote(alertid, note string) (string, error) {
	data := fmt.Sprintf("{\"source\":\"opsgenie-processor\",\"note\":\"%v\"}", note)
	results, err := opsgenieAPI(alertid, "notes", data)
	return fmt.Sprintf("OpsGenie ACTION CALLED: Note() %v", results), err
}

// call the OpsGenie API
func opsgenieAPI(alertid, identifier, payload string) (string, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://api.opsgenie.com/v2/alerts/"+alertid+"/"+identifier, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return fmt.Sprintf("opsgenieAPI(%v): http.NewRequest err = %v", alertid, err.Error()), err
	}

	req.Header.Add("Authorization", "GenieKey "+os.Getenv("OPSGENIE_KEY"))
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("opsgenieAPI(%v): client.Do err = %v", alertid, err.Error()), err
	}
	defer resp.Body.Close()

	return fmt.Sprintf("opsgenieAPI(%v): resp.StatusCode = %v", alertid, resp.StatusCode), nil
}

// buildArray constructs up a JSON array
func buildArray(items []string) string {
	line := "["
	for i, v := range items {
		if i > 0 {
			line = line + ","
		}
		line = line + fmt.Sprintf("\"%v\"", v)
	}
	line = line + "]"
	return line
}
