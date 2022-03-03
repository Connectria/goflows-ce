// sms.go - For sending sms messages via twilio
package goflows

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

var SkylineSMS = map[string]interface{}{
	"SendSMS": sendSMS,
}

type SMSMessage struct {
	Message string `json:"body"`
}

func sendSMS(to_number string, message string) string {

	sms := new(SMSMessage)
	sms.Message = message

	payload, err := json.Marshal(sms)
	if err != nil {
		return "I had an issue with the Marshal of SMS"
	}

	// TODO: Remove SKYLINE_SMS_URL
	resp, _ := http.Post(os.Getenv("SKYLINE_SMS_URL")+"/sms/"+to_number+"/send/sms",
		"application/json",
		bytes.NewBuffer(payload),
	)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Sorry, but an error occurred while attempting to read response body Error:" + err.Error()
	}
	defer resp.Body.Close()

	return string(body)
}
