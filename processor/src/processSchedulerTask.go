// processSchedulerTask.go - execute GoFlow specified by scheduler

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"goflows-processor/flows"
	"goflows-processor/flows/goflows"
)

type schedulerType struct {
	TriggerID   string `json:"triggerId"`
	TriggerName string `json:"triggerName"`
	FlowIDs     []struct {
		FlowID          string `json:"flowID"`
		FuncName        string `json:"funcName"`
		FlowDescription string `json:"flowDescription"`
		FlowName        string `json:"flowName"`
	} `json:"flowIDs"`
	AtTime      int64                   `json:"atTime,omitempty"`
	DutyCycle   string                  `json:"dutyCycle,omitempty"`
	RepeatUntil int64                   `json:"repeat-until,omitempty"`
	Inputs      []goflows.FlowInputType `json:"inputs"`
}

// processSchedulerTask execute GoFlow specified by scheduler
func processSchedulerTask(rt string) error {

	var taskFlow = func(*goflows.GoFlow) bool { return false }

	var receivedTask schedulerType
	if err := json.Unmarshal([]byte(rt), &receivedTask); err != nil {
		err := errors.New("Unable to unmarshal data received from scheduler: " + fmt.Sprintf("%v", rt))
		logger.Error().
			Str("function", "processSchedulerTask()").
			Msg(err.Error())
		return err
	}

	logger.Info().
		Str("function", "processSchedulerTask()").
		Str("triggerId", receivedTask.TriggerID).
		Str("triggerName", receivedTask.TriggerName).
		Msg("GoFlow task received from scheduler.")

	// look through the provided flows, if ANY not found, error
	found := 0
	for _, providedFlow := range receivedTask.FlowIDs {
		for _, flow := range flows.TaskFlowList {
			funcName := flows.GetFuncName(flow)
			if providedFlow.FuncName == funcName {
				found++
				break
			}
		}
	}

	if found != len(receivedTask.FlowIDs) {
		err := errors.New("required GoFlows not found in TaskFlowList for TriggerID")
		logger.Error().
			Str("function", "processSchedulerTask()").
			Str("triggerId", receivedTask.TriggerID).
			Msg(err.Error())
		return err
	}

	// process the trigger which may be multiple funcs
	startTrigger := time.Now()
	logger.Info().
		Str("function", "processSchedulerTask()").
		Str("triggerId", receivedTask.TriggerID).
		Msg("Initiating TriggerID.")

	for i, providedFlow := range receivedTask.FlowIDs {
		gf := goflows.New("TASK: "+providedFlow.FuncName, &logger, mongoClient, cfg.AMQjobStatus)
		logger.Info().
			Str("function", "processSchedulerTask()").
			Str("triggerId", receivedTask.TriggerID).
			Str("jobID", gf.JobID).
			Msg("GoFlow assigned jobID.")

		// setup - TriggerID - may be redundant?
		if len(receivedTask.TriggerID) > 0 {
			gf.TriggerID = receivedTask.TriggerID
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided TriggerID: %v", gf.TriggerID)
		}

		// setup - common used vars
		// Commented out by AKB for CE release
		//gf.SetFlowVar("EVENTBRIDGE_SOURCE", cfg.EventBridgeSource)

		/**
		// TODO: Ask VS why this is in the processSchedulerTask
		// This statement is simular as in processOpsGenieEvent.go
		// While in there it would make sense, as we are processing OpsGenieEvents
		//  I am unclear why its in here (unless at this glace my underingstanding is off)
		**/
		//gf.SetFlowVar("OPSGENIE_TAG", cfg.OpsGenieTag)
		//gf.Tags = append(gf.Tags, cfg.OpsGenieTag)

		// setup - Debug
		gf.Debug = false

		// setup - FuncName
		if len(providedFlow.FlowName) > 0 {
			gf.FuncName = providedFlow.FuncName
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided FuncName: %v", gf.FuncName)
		}

		// setup - JobFlowInputs
		if len(receivedTask.Inputs) > 0 {
			gf.JobFlowInputs = receivedTask.Inputs
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided JobFlowInputs: %v", gf.JobFlowInputs)
		}

		// setup - JobName
		if len(receivedTask.TriggerName) > 0 {
			gf.JobName = receivedTask.TriggerName
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided JobName: %v", gf.JobName)
		}

		// setup - JobSrcFlowID
		if len(providedFlow.FlowID) > 0 {
			//	gf.JobSrcFlowID = providedFlow.FlowID
			gf.JobSrcFlowID = receivedTask.FlowIDs[i].FlowID
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("providedFlow.FlowID='%v' - gf.JobSrcFlowID='%v'", receivedTask.FlowIDs[i].FlowID, gf.JobSrcFlowID)
		}

		// setup - FlowName
		if len(providedFlow.FlowName) > 0 {
			gf.JobSrcFlowName = providedFlow.FlowName
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided FlowName: %v", gf.JobSrcFlowName)
		}

		// setup - FlowDescription
		if len(providedFlow.FlowDescription) > 0 {
			gf.FlowDescription = providedFlow.FlowDescription
			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msgf("Provided FlowDescription: %v", gf.FlowDescription)
		}

		// send initial status
		gf.UpdateStatus(
			"Submitted", // runStatus
			"",          // stepStatus
			"",          // info
			0.0,         // duration
		)

		// find the func to execute
		for _, flow := range flows.TaskFlowList {
			funcName := flows.GetFuncName(flow)
			if providedFlow.FuncName == funcName {
				taskFlow = flow
				break
			}
		}

		// execute the function
		startTime := time.Now()
		status := taskFlow(gf)
		duration := time.Since(startTime).Seconds()
		logger.Info().
			Str("function", "processSchedulerTask()").
			Str("triggerId", receivedTask.TriggerID).
			Str("jobID", gf.JobID).
			Msgf("Returned: %v", status)

		// no errors
		gf.JobControlExit = 0

		// action taken
		if status {
			if gf.Error {
				gf.JobControlExit = 1 // errors occurred
			}

			gf.UpdateStatus(
				"Success",                           // runStatus
				"",                                  // stepStatus
				"Action taken by '"+gf.FuncName+"'", // info
				duration,                            // duration
			)

			logger.Info().
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msg("Returned true.")
		}

		// no action taken
		if gf.JobStepID == 0 {
			if gf.Error {
				gf.JobControlExit = 1 // errors occurred
			}

			gf.UpdateStatus(
				"Finished",                       // runStatus
				"",                               // stepStatus
				"No action by '"+gf.FuncName+"'", // info
				duration,                         // duration
			)

			logger.Info().
				Float64("duration", duration).
				Str("function", "processSchedulerTask()").
				Str("triggerId", receivedTask.TriggerID).
				Str("jobID", gf.JobID).
				Msg("Finished GoFlow.")
		}

		/** Disable Callbacks
				gf.RemoveCallBack()
		**/
		gf.CloseJobsStatusQ()
	}

	logger.Info().
		Float64("duration", time.Since(startTrigger).Seconds()).
		Str("function", "processSchedulerTask()").
		Str("triggerId", receivedTask.TriggerID).
		Msg("Finished TriggerID.")

	return nil
}
