package goflows

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateStatus collection for JobID/StepID
func (gf *GoFlow) UpdateStatus(runStatus, stepStatus, info string, duration float64) {

	// common time for mongodb and rabbitMQ
	now := time.Now()
	updateTime := now.Unix()
	updateAtTime := primitive.DateTime(now.UnixNano() / int64(time.Millisecond))

	// mongodb
	err := updateStatusMongo(statusDataType{
		JobControlExit:   gf.JobControlExit,
		JobControlStatus: runStatus,
		JobCreateTime:    gf.JobCreateTime,
		JobDuration:      duration,
		JobFlowInputs:    gf.JobFlowInputs,
		JobID:            gf.JobID,
		JobInfo:          info,
		JobName:          gf.JobName,
		JobSrcFlowID:     gf.JobSrcFlowID,
		JobSrcFlowName:   gf.JobSrcFlowName,
		JobStepID:        gf.JobStepID,
		JobStepStatus:    stepStatus,
		JobStatusTime:    updateTime,
		TriggerID:        gf.TriggerID,
		UpdatedAt:        updateAtTime,
	})

	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Int("jobStepID", gf.JobStepID).
			Msgf("updateStatusMongo() returned error: '%v'", err.Error())
	}

	// Write status information to RabbitMQ Exchange
	postBody, _ := json.Marshal(statusDataType{
		JobControlExit:   gf.JobControlExit,
		JobControlStatus: runStatus,
		JobCreateTime:    gf.JobCreateTime,
		JobDuration:      duration,
		JobFlowInputs:    gf.JobFlowInputs,
		JobID:            gf.JobID,
		JobInfo:          info,
		JobName:          gf.JobName,
		JobSrcFlowID:     gf.JobSrcFlowID,
		JobSrcFlowName:   gf.JobSrcFlowName,
		JobStepID:        gf.JobStepID,
		JobStepStatus:    stepStatus,
		JobStatusTime:    updateTime,
		TriggerID:        gf.TriggerID,
		UpdatedAt:        updateAtTime,
	})

	err = gf.updateJobStatusQ([]byte(postBody))
	if err != nil {
		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Int("jobStepID", gf.JobStepID).
			Msgf("updateJobsStatusQ() returned error: '%v'", err.Error())
	}
}
