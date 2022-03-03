// slack.go - Post to slack using a provided webhook
package goflows

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var Slack = map[string]interface{}{
	"PostSlack": postSlack,
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

// postSlack will post a message to an 'Incoming Webook' url setup in Slack Apps
func postSlack(webhook string, message string) error {

	slackBody, _ := json.Marshal(SlackRequestBody{Text: message})
	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return fmt.Errorf("non-ok response returned from Slack")
	}

	return nil
}
