// callback.go 	- GoFlow callback handling

package goflows

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type CBInputElements struct {
	InputName  string      `json:"inputName"`
	InputValue interface{} `json:"inputValue"`
}

type CallBackInputSlice struct {
	Inputs []CBInputElements `json:"inputs"`
}

// TODO: Understand Why PortalContactUser is in callback, and if its in later versions
type PortalContactUser struct {
	ID    string
	Name  string
	Email string
}

type WebElements struct {
	Note            string
	TriggerContacts []PortalContactUser
	PortalContacts  []PortalContactUser
	PortalUsers     []PortalContactUser
}

const (
	CB_STATUS_NEEDED     = "needed"
	CB_STATUS_TAKEN      = "taken"
	CB_STATUS_PROCESSING = "processing"
	CB_STATUS_EMPTY      = "not-needed"
)

const (
	AA_ACTION_EXPIRED = "ALERT-ACTION-EXPIRED"
	AA_ACTION_TAKEN   = "ALERT-ACTION-TAKEN"
	AA_NOACTION_TAKEN = "ALERT-NOACTION-TAKEN"
)

// AlertActionType
type AlertActionType struct {
	AlertID  string      `bson:"alertID,omitempty",json:"alerID,omitempty"`          // OpsGenie alert needing action
	Status   string      `bson:"status,omitempty",json:"status,omitempty"`           // see const above
	Elements WebElements `bson:"webElements,omitempty",json:"webElements,omitempty"` // provided to help user take action
}

// CallBackType holds call back data
type CallBackType struct {
	JobID       string             `bson:"jobID",json:"jobID"`
	Reference   string             `bson:"callBackRef",json:"callBackRef"`
	ID          string             `bson:"callBackID",json:"callBackID"`
	Data        CallBackInputSlice `bson:"callBackData",json:"callBackData"`
	AlertAction AlertActionType    `bson:"alertAction,omitempty",json:"alertAction,omitempty"`
}

// lookup alert action and status
func LookupCallBackAlertID(alertID string) string {
	pipeline := []bson.D{
		bson.D{
			{"$match", bson.D{
				{"alertAction.AlertID", alertID},
			}},
		},
	}

	callBacks := GetCallBacks(gfMongoClient, pipeline)
	for _, v := range callBacks {
		if alertID == v.AlertAction.AlertID {
			if v.AlertAction.Status == "" {
				return CB_STATUS_EMPTY
			}
			return v.AlertAction.Status
		}
	}

	// not found, so not-needed
	return CB_STATUS_EMPTY
}

// update the call back alert status
// TODO: More Portal stuff, due to AlertActions
func (gf *GoFlow) UpdateCallBackStatusForAlertID(ref, status string) {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("UpdateCallBackStatusForAlertID(ref = '%v'; status = '%v')", ref, status)
	}

	// lookup and update call back
	callBack, _ := getCallBackByReference(gfMongoClient, ref)
	newCallBack := CallBackType{
		JobID:     gf.JobID,
		Reference: ref,
		ID:        callBack.ID,
		Data:      callBack.Data,
		AlertAction: AlertActionType{
			AlertID: callBack.AlertAction.AlertID,
			Status:  status,
			Elements: WebElements{
				Note:            callBack.AlertAction.Elements.Note,
				TriggerContacts: callBack.AlertAction.Elements.TriggerContacts,
				PortalContacts:  callBack.AlertAction.Elements.PortalContacts,
				PortalUsers:     callBack.AlertAction.Elements.PortalUsers,
			},
		},
	}

	//update mongo collection
	err := UpdateCallBackMongo(gfMongoClient, newCallBack)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("UpdateCallBackStatusForAlertID('%v', '%v') updateCallBackMongo returned error - %v", ref, status, err.Error())
		return
	}

	gf.FlowLogger.Info().
		Str("jobID", gf.JobID).
		Msgf("UpdateCallBackStatusForAlertID('%v', '%v') updated: %v", ref, status, callBack)
}

// create the call back alert
func (gf *GoFlow) CreateCallBackForAlertID(ref, id, status, note, triggerContacts, customerID string) {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("CreateCallBackForAlertID(ref = '%v'; id = '%v'; status = '%v'; triggerContacts = '%v', customerID = '%v')", ref, id, status, triggerContacts, customerID)
	}

	// pull contacts by customer
	contacts, err := CustomerUsers(customerID, true)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("CreateCallBackForAlertID('%v') CustomerContacts('%v', true) returned error - %v", ref, customerID, err.Error())
		return
	}

	// contacts provided by event trigger
	var tc []PortalContactUser
	if triggerContacts != KEYNOTFOUND && len(triggerContacts) > 0 {
		for _, w := range strings.Split(triggerContacts, ",") {
			if contacts.Response.Status {
				for _, x := range contacts.Response.Data.Users {
					if gf.Debug {
						gf.FlowLogger.Debug().
							Str("jobID", gf.JobID).
							Msgf("CreateCallBackForAlertID(ref = '%v') Checking event trigger contact '%v' for selection: active = '%v'; email = '%v'", ref, x.ID, x.Active, x.Email)
					}

					if (strconv.Itoa(x.ID) == w) && x.Active {
						tc = append(
							tc, PortalContactUser{
								ID:    strconv.Itoa(x.ID),
								Name:  x.Fname + " " + x.Surname,
								Email: x.Email,
							})
					}
				}
			}
		}
	}

	// portal contacts for customer
	var pc []PortalContactUser
	for _, x := range contacts.Response.Data.Users {
		if gf.Debug {
			gf.FlowLogger.Debug().
				Str("jobID", gf.JobID).
				Msgf("CreateCallBackForAlertID(ref = '%v') Checking portal contact '%v' to build selection: active = '%v'; email = '%v'", ref, x.ID, x.Active, x.Email)
		}

		if x.Active {
			pc = append(
				pc, PortalContactUser{
					ID:    strconv.Itoa(x.ID),
					Name:  x.Fname + " " + x.Surname,
					Email: x.Email,
				})
		}
	}

	// pull users by customer
	users, err := CustomerUsers(customerID, false)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("CreateCallBackForAlertID('%v') CustomerUsers('%v') returned error - %v", ref, customerID, err.Error())
		return
	}

	var pu []PortalContactUser
	for _, x := range users.Response.Data.Users {
		if x.Active {
			pu = append(
				pu, PortalContactUser{
					ID:    strconv.Itoa(x.ID),
					Name:  x.Fname + " " + x.Surname,
					Email: x.Email,
				})
		}
	}

	// lookup and update call back
	callBack, _ := getCallBackByReference(gfMongoClient, ref)
	newCallBack := CallBackType{
		JobID:     gf.JobID,
		Reference: ref,
		ID:        id,
		Data:      callBack.Data,
		AlertAction: AlertActionType{
			AlertID: id,
			Status:  status,
			Elements: WebElements{
				Note:            note,
				TriggerContacts: tc,
				PortalContacts:  pc,
				PortalUsers:     pu,
			},
		},
	}

	//update mongo collection
	err = UpdateCallBackMongo(gfMongoClient, newCallBack)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("CreateCallBackForAlertID('%v') updateCallBackMongo returned error - %v", ref, err)
		return
	}

	gf.FlowLogger.Info().
		Str("jobID", gf.JobID).
		Msgf("CreateCallBackForAlertID('%v') updated: %v", ref, callBack)
}

// GenerateCallBackURL returns the URL for external system to POST to running GoFlow
func (gf *GoFlow) GenerateCallBackURL(ref string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GenerateCallBackURL(ref = '%v')", ref)
	}

	// generate ID
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("GenerateCallBackURL(ref = '%v') error generating callBackID: %v", ref, err.Error())
		return ""
	}

	ID := fmt.Sprintf("%x-%x", b[0:4], b[4:])
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GenerateCallBackURL(%v) created call back reference: %v", ref, ID)
	}

	// format URL
	callBackURL := fmt.Sprintf("%v/%v", os.Getenv("CALLBACK_PREFIX_URL"), ID)
	gf.FlowLogger.Info().
		Str("jobID", gf.JobID).
		Msgf("GenerateCallBackURL(%v) generated: %v", ref, callBackURL)

	// create new entry
	callBack := CallBackType{
		JobID:     gf.JobID,
		Reference: ref,
		ID:        ID,
		Data:      CallBackInputSlice{},
	}

	err = UpdateCallBackMongo(gfMongoClient, callBack)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("GenerateCallBackURL(%v): updateCallBackMongo returned error - %v", ref, err.Error)
	}

	gf.FlowLogger.Info().
		Str("jobID", gf.JobID).
		Msgf("GenerateCallBackURL(%v) created new entry: %v", ref, callBack)

	// create Kong service
	err = kongAddService(ID)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("kongAddService(%v) generated: %v", ID, err.Error())
		gf.Error = true
	} else {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msgf("Kong service added for %v", ID)
	}

	// add route to Kong
	err = kongAddRoute(ID)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("kongAddRoute(%v) generated: %v", ID, err.Error())
		gf.Error = true
	} else {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msgf("Kong route added for %v", ID)
	}

	// return the call back URL
	return callBackURL
}

// WaifForCallBack looks for POST data
func (gf *GoFlow) WaifForCallBack(ref string, secs int) bool {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("WaitForCallBack(ref = '%v', secs = '%v')", ref, secs)
	}

	// set the timer
	waitTime := time.Duration(secs) * time.Second

	gf.FlowLogger.Info().
		Str("jobID", gf.JobID).
		Msgf("WaitForCallBack(%v, %v)", ref, waitTime)

	gf.UpdateStatus(
		"Running", // runStatus
		"Waiting", // stepStatus
		"Waiting for call back data to be received", // info
		0.0, // duration
	)

	time.Sleep(waitTime)

	// lookup call back
	callBack, err := getCallBackByReference(gfMongoClient, ref)
	if err != nil {
		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Msgf("WaitForCallBack(%v, %v) invalid reference; returning false", ref, waitTime)

		gf.UpdateStatus(
			"Running", // runStatus
			"Timeout", // stepStatus
			"GoFlow contains invalid call back reference: "+ref, // info
			float64(secs), // duration
		)

		gf.Error = true
		return false
	}

	if len(callBack.Data.Inputs) > 0 {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msgf("WaitForCallBack(%v, %v) found POST data( %v ); returning true", ref, waitTime, callBack.Data.Inputs)

		gf.UpdateStatus(
			"Running", // runStatus
			"Running", // stepStatus
			fmt.Sprintf("Call back data received: %v", callBack.Data.Inputs), // info
			float64(secs), // duration
		)

		return true
	}

	gf.FlowLogger.Warn().
		Str("jobID", gf.JobID).
		Msgf("WaitForCallBack(%v, %v) did not find POST data; returning false", ref, waitTime)

	gf.UpdateStatus(
		"Running",                                // runStatus
		"Timeout",                                // stepStatus
		"Call back timed out, data not recieved", // info
		float64(secs),                            // duration
	)

	return false
}

// GetCallBackInput returns the value by the supplied key (assuming it was in the POST data)
func (gf *GoFlow) GetCallBackInput(ref, key string) interface{} {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GetCallBackInput(ref = %v, key = '%v')", ref, key)
	}

	// lookup call back
	callBack, err := getCallBackByReference(gfMongoClient, ref)
	if err != nil {
		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Msgf("GetCallBackInput(): call back reference ('%v') does not exist", ref)
		return nil
	}

	// return the key value
	for _, v := range callBack.Data.Inputs {
		if v.InputName == key {
			return v.InputValue
		}
	}

	// nothing found (i.e., ref or key invalid)
	return nil
}

// GetCallBackPretty returns the value of a callback in pretty format
func (gf *GoFlow) GetCallBackPretty(ref string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GetCallBackPretty(ref = '%v')", ref)
	}

	// lookup call back
	callBack, err := getCallBackByReference(gfMongoClient, ref)
	if err != nil {
		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Msg("GetCallBackPretty(): call back reference does not exist")
		return "call back reference does not exist"
	}

	var prettyJSON bytes.Buffer
	b, _ := json.Marshal(callBack.Data)
	_ = json.Indent(&prettyJSON, b, "", "    ")
	return prettyJSON.String()
}

// RemoveCallBack removes a call back
func (gf *GoFlow) RemoveCallBack() {

	// lookup call back
	callBack, err := getCallBackByJobID(gfMongoClient, gf.JobID)
	if err != nil {
		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Msg("RemoveCallBack(): JobID does not have a call back")
		return
	}

	// delete call back from collection
	err = DeleteCallBackByJobID(gfMongoClient, gf.JobID)
	if err != nil {
		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Msgf("RemoveCallBack(): deleteCallBackByJobID() returned an error - %v", err.Error)
	} else {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msg("RemoveCallBack()")
	}

	// remove route from Kong
	err = kongRemoveRoute(callBack.ID)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("kongRemoveRoute(%v) generated: %v", callBack.ID, err.Error())
		gf.Error = true
	} else {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msgf("RemoveCallback(): Kong route removed for %v", callBack.ID)
	}

	// remove service from Kong
	err = kongRemoveService(callBack.ID)
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Msgf("kongRemoveService(%v) generated: %v", callBack.ID, err.Error())
		gf.Error = true
	} else {
		gf.FlowLogger.Info().
			Str("jobID", gf.JobID).
			Msgf("Kong service removed for %v", callBack.ID)
	}
}

// Kong
type newKongService struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// add a Kong service to match internal call back
func kongAddService(name string) error {

	// check to see if the service already exists
	status, err := kongAdminAPI("GET", "/services/callback/"+name, nil)
	if err != nil {
		return err
	}

	// found, so delete, since the config (i.e., port) may have changed
	if status == 200 {
		err = kongRemoveService(name)
		if err != nil {
			return err
		}
	}

	// create the service
	data := newKongService{
		Name: name,
		URL:  os.Getenv("CALLBACK_INTERNAL_URL") + "/callback/" + name,
	}

	d, _ := json.Marshal(data) // convert struct to []byte
	status, err = kongAdminAPI("POST", "/services/", d)
	if err != nil {
		return err
	}

	if status != 201 { // status is not created
		return fmt.Errorf("kongAdminAPI(%v, %v, %v) returned %v", "POST", "/services/", data, status)
	}

	return nil
}

// remove the Kong service
func kongRemoveService(name string) error {

	// TODO: if there are any routes left (due to crashes?), remove those before removing service otherwise this will fail

	_, err := kongAdminAPI("DELETE", "/services/"+name, nil)
	if err != nil {
		return err
	}

	return nil
}

type newKongRoute struct {
	Methods []string `json:"methods"`
	Name    string   `json:"name"`
	Paths   []string `json:"paths"`
}

// add route to Kong - for temporary call back
func kongAddRoute(name string) error {
	data := newKongRoute{
		Methods: []string{"POST"},
		Name:    name,
		Paths:   []string{"/" + name},
	}

	d, _ := json.Marshal(data) // convert struct to []byte
	status, err := kongAdminAPI("POST", "/services/"+name+"/routes", d)
	if err != nil {
		return err
	}

	if status != 201 {
		return fmt.Errorf("kongAdminAPI(%v, %v, %v) returned %v", "POST", "/services/"+name+"/routes", data, status)
	}

	return nil
}

// remove route
func kongRemoveRoute(name string) error {
	_, err := kongAdminAPI("DELETE", "/services/"+name+"/routes/"+name, nil)
	if err != nil {
		return err
	}

	return nil
}

// call the Kong Admin API
func kongAdminAPI(method, directive string, payload []byte) (int, error) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}

	combinedURL := os.Getenv("KONG_ADMIN_API_URL") + directive
	req, err := http.NewRequest(method, combinedURL, bytes.NewBuffer(payload))
	if err != nil {
		return -1, err
	}

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()
	return resp.StatusCode, nil
}
