// dbstatus.go -  goflows status collection

package goflows

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type statusDataType struct {
	JobControlExit   int64              `json:"jobControlExit" bson:"jobControlExit"`
	JobControlStatus string             `json:"jobControlStatus" bson:"jobControlStatus"`
	JobCreateTime    int64              `json:"jobCreateTime" bson:"jobCreateTime"`
	JobDuration      float64            `json:"jobDuration" bson:"jobDuration"`
	JobFlowInputs    []FlowInputType    `json:"jobFlowInputs" bson:"jobFlowInputs"`
	JobID            string             `json:"jobID" bson:"jobID"`
	JobInfo          string             `json:"jobInfo" bson:"jobInfo"`
	JobName          string             `json:"jobName" bson:"jobName"`
	JobSrcFlowID     string             `json:"jobSrcFlowID" bson:"jobSrcFlowID"`
	JobSrcFlowName   string             `json:"jobSrcFlowName" bson:"jobSrcFlowName"`
	JobStepID        int                `json:"jobStepID,omitempty" bson:"jobStepID"`
	JobStepStatus    string             `json:"jobStepStatus,omitempty" bson:"jobStepStatus"`
	JobStatusTime    int64              `json:"jobStatusTime" bson:"jobStatusTime"`
	TriggerID        string             `json:"triggerId" bson:"triggerId"`
	UpdatedAt        primitive.DateTime `json:"updatedAt" bson:"updatedAt"`
}

func updateStatusMongo(rec statusDataType) error {

	statusCollection := gfMongoClient.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_STATUS_COLLECTION"))
	opts := options.Update().SetUpsert(true)

	var filter, update bson.M
	filterString := fmt.Sprintf("{\"jobID\": \"%v\"}", rec.JobID)
	_ = bson.UnmarshalExtJSON([]byte(filterString), true, &filter)

	update = bson.M{
		"$set": bson.M{
			"jobControlExit":   rec.JobControlExit,
			"jobControlStatus": rec.JobControlStatus,
			"jobCreateTime":    rec.JobCreateTime,
			"jobDuration":      rec.JobDuration,
			"jobFlowInputs":    rec.JobFlowInputs,
			"jobID":            rec.JobID,
			"jobInfo":          rec.JobInfo,
			"jobName":          rec.JobName,
			"jobSrcFlowID":     rec.JobSrcFlowID,
			"jobSrcFlowName":   rec.JobSrcFlowName,
			"jobStepID":        rec.JobStepID,
			"jobStepStatus":    rec.JobStepStatus,
			"jobStatusTime":    rec.JobStatusTime,
			"triggerId":        rec.TriggerID,
			"updatedAt":        rec.UpdatedAt,
		},
	}

	_, err := statusCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

type countStatusType struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Count int64              `bson:"count,omitempty"`
}

// return information from status needed for statistic functions
func countStatusMongo(p []primitive.D) int64 {
	statusCollection := gfMongoClient.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_STATUS_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := statusCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countStatusMongo() statusCollection.Aggregate returned error = %v\n", err.Error())
		return -1
	}

	var info []countStatusType
	err = loadedStructCursor.All(context.TODO(), &info)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countStatusMongo() loadedStructCursor.All returned error = %v\n", err.Error())
		return -1
	}

	if len(info) == 0 {
		return 0
	}

	return info[0].Count
}
