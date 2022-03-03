// processOpsGenieEvent.go - evaluate an OpsGenie event against event triggers

package main

import (
	"fmt"
	"time"

	"goflows-processor/flows"
	"goflows-processor/flows/goflows"

	"github.com/tidwall/gjson"
)

// processOpsGenieEvent evaluates an OpsGenie event against event triggers
func processOpsGenieEvent(messageID string) error {
	logger.Info().
		Str("function", "processOpsGenieEvent()").
		Str("messageID", messageID).
		Msg("messaged received")

	// retreive event details
	thisEvent, err := lookupMessageID(messageID)
	if err != nil {
		logger.Error().
			Str("function", "processOpsGenieEvent()").
			Str("messageID", messageID).
			Msgf("lookupMessageID returned: '%v'", err.Error())
		return err
	}

	// unknown event action
	if len(getOpsGenieEventFieldValue(thisEvent, "EventData.action")) == 0 {
		logger.Warn().
			Str("function", "processOpsGenieEvent()").
			Str("messageID", messageID).
			Msg("empty OpsGenie event")
		return nil
	}

	// set the alertId used in logging
	thisAlertID := getOpsGenieEventFieldValue(thisEvent, "EventData.alert.alertId")

	// only evaluating new OpsGenie events
	if getOpsGenieEventFieldValue(thisEvent, "EventData.action") != "Create" {
		logger.Info().
			Str("function", "processOpsGenieEvent()").
			Str("messageID", messageID).
			Str("eventAlertID", thisAlertID).
			Msg("not a new event; skipping")
		return nil
	}

	// create a new job to run through the event triggers
	gf := goflows.New(messageID, &logger, mongoClient, cfg.AMQjobStatus)
	logger.Info().
		Str("function", "processOpsGenieEvent()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Str("jobID", gf.JobID).
		Msg("assigned jobID")

	// set so Flows can utilize
	// TODO: Eventbridge Referance but not a show stapper for publishing to public repo
	gf.SetFlowVar("eventAlertID", thisAlertID)
	gf.SetFlowVar("eventAlertAlias", getOpsGenieEventFieldValue(thisEvent, "EventData.alert.alias"))

	//Commented out by AKB for CE release
	//gf.SetFlowVar("EVENTBRIDGE_SOURCE", cfg.EventBridgeSource)

	/**
	// The OpsGenieTag is set in the config.go and reads OPSGENIE_TAG,
	// This is the default tag that is added to an OPSGENIE alert that is processed
	// (AKB) While I am ok with us bringing in OpsGenie events, I guess I am less ok with
	//  forcing a "tag-back", So I am going to see if I can extract this.
	// If I am reading this correctly, its only being added here, I have not confirmed where
	//  the tag-back is being preformed but we would look for a bulk gf.Tags if this is the case?
	**/
	//gf.SetFlowVar("OPSGENIE_TAG", cfg.OpsGenieTag)
	//gf.Tags = append(gf.Tags, cfg.OpsGenieTag)

	// if there are any alert tags, add them for consistency1
	tags := getOpsGenieEventFieldBytes(thisEvent, "EventData.alert.tag")
	if tags.Index > 0 {
		tags.ForEach(
			func(k, v gjson.Result) bool {
				gf.Tags = append(gf.Tags, v.String())
				return true
			})
	}

	// OpsGenie alert time for use in flows - yes, it's a temporary hack - seriously
	gf.SetFlowVar("eventTime",
		convertCreateAtStr(
			getOpsGenieEventFieldValue(thisEvent, "EventData.alert.createdAt"),
		),
	)

	logger.Info().
		Str("function", "processOpsGenieEvent()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Str("jobID", gf.JobID).
		Msgf("eventTime = '%v'", gf.GetFlowVar("eventTime"))

	// starting event analysis
	startTime := time.Now()
	logger.Info().
		Str("function", "processOpsGenieEvent()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Str("jobID", gf.JobID).
		Msg("starting trigger event rule analysis")

	//  evaluate rules; no errors == rule was matched.
	evalResp, err := evalRules(messageID)
	if err == nil {

		// update the jobID with the appropriate triggerID
		gf.TriggerID = evalResp.trigger.TriggerID

		// pass any flowvars that were created during the evalRules()
		for k, v := range evalResp.setVars {
			gf.SetFlowVar(k, v)
		}

		// send initial status
		startTime := time.Now()
		logMsg := "all event rules (criteria) met; initiating trigger"
		gf.UpdateStatus(
			"Submitted", // runStatus
			"",          // stepStatus
			logMsg,      // info
			0.0,         // duration
		)
		logger.Info().
			Str("function", "processOpsGenieEvent()").
			Str("messageID", messageID).
			Str("triggerId", gf.TriggerID).
			Str("eventAlertID", thisAlertID).
			Msg(logMsg)

		// loop through list of flows provided in trigger
		for _, flowID := range evalResp.trigger.Flowids {
			gf.JobSrcFlowID = flowID

			// retrieve flow from front end API
			f, err := goflows.GetFlow(flowID)
			if err != nil {
				logMsg = fmt.Sprintf("goflows.GetFlow('%v') returned error: '%v'", flowID, err.Error())
				logger.Error().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("triggerId", gf.TriggerID).
					Str("eventAlertID", thisAlertID).
					Msg(logMsg)
				gf.Error = true
				break
			}

			// setup - FlowName
			if len(f.FlowName) > 0 {
				gf.JobSrcFlowName = f.FlowName
				logger.Info().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("triggerId", gf.TriggerID).
					Str("eventAlertID", thisAlertID).
					Str("jobID", gf.JobID).
					Msgf("Provided FlowName: %v", gf.JobSrcFlowName)
			}

			// setup - FlowDescription
			if len(f.FlowDescription) > 0 {
				gf.FlowDescription = f.FlowDescription
				logger.Info().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("triggerId", gf.TriggerID).
					Str("eventAlertID", thisAlertID).
					Str("jobID", gf.JobID).
					Msgf("Provided FlowDescription: %v", gf.FlowDescription)
			}

			// setup - JobFlowInputs
			if len(evalResp.trigger.Inputs) > 0 {
				gf.JobFlowInputs = evalResp.trigger.Inputs
				logger.Info().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("triggerId", gf.TriggerID).
					Str("eventAlertID", thisAlertID).
					Str("jobID", gf.JobID).
					Msgf("provided JobFlowInputs: %v", gf.JobFlowInputs)
			}

			// setup - JobName
			if len(evalResp.trigger.Name) > 0 {
				gf.JobName = evalResp.trigger.Name
				logger.Info().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("triggerId", gf.TriggerID).
					Str("eventAlertID", thisAlertID).
					Str("jobID", gf.JobID).
					Msgf("Provided JobName: %v", gf.JobName)
			}

			// find the func to execute
			var taskFlow = func(*goflows.GoFlow) bool { return false }
			for _, flow := range flows.TaskFlowList {
				funcName := flows.GetFuncName(flow)
				if f.FlowDocument.FuncName == funcName {
					gf.FuncName = funcName
					taskFlow = flow
					break
				}
			}

			// setup - always start with debug off
			gf.Debug = false

			// execute the function
			status := taskFlow(gf)
			logger.Info().
				Str("function", "processOpsGenieEvent()").
				Str("messageID", messageID).
				Str("triggerId", gf.TriggerID).
				Str("eventAlertID", thisAlertID).
				Str("jobID", gf.JobID).
				Msgf("returned: %v", status)

			// no errors
			gf.JobControlExit = 0
			logMsg := fmt.Sprintf("action taken by '%v'", gf.FuncName)
			if gf.Error {
				gf.JobControlExit = 1 // errors occurred
				logMsg = fmt.Sprintf("%v; check logs for errors", logMsg)
			}

			duration := time.Since(startTime).Seconds()
			gf.UpdateStatus(
				"Success", // runStatus
				"",        // stepStatus
				logMsg,    // info
				duration,  // duration
			)
			logger.Info().
				Str("function", "processOpsGenieEvent()").
				Str("messageID", messageID).
				Str("triggerId", gf.TriggerID).
				Str("eventAlertID", thisAlertID).
				Str("jobID", gf.JobID).
				Msg(logMsg)
		}

		// no action taken
		if gf.JobStepID == 0 {
			logMsg := "no action taken"
			if gf.Error {
				gf.JobControlExit = 1 // errors occurred
				logMsg = "check logs for errors"
			}

			duration := time.Since(startTime).Seconds()
			gf.UpdateStatus(
				"Finished", // runStatus
				"",         // stepStatus
				logMsg,     // info
				duration,   // duration
			)
			logger.Info().
				Float64("duration", duration).
				Str("function", "processOpsGenieEvent()").
				Str("messageID", messageID).
				Str("eventAlertID", thisAlertID).
				Str("triggerId", gf.TriggerID).
				Str("jobID", gf.JobID).
				Msg(logMsg)
		}
	}

	// finished
	duration := time.Since(startTime).Seconds()
	logger.Info().
		Float64("duration", duration).
		Str("function", "processOpsGenieEvent()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Str("jobID", gf.JobID).
		Msg("finished event trigger analysis")

	// cleanup
	/** Disable Callbacks
		gf.RemoveCallBack()
	**/
	gf.CloseJobsStatusQ()
	return nil
}
