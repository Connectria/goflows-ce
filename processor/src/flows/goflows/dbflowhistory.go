// dbflowhistory.go -  goflows history collection

package goflows

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type countFlowHistoryType struct {
	Count int64 `bson:"count,omitempty"`
}

// return count from log based on p for statistic functions
func countFlowHistoryMongo(p []primitive.D) int64 {
	flowHistoryCollection := gfMongoClient.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_LOG_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := flowHistoryCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countFlowHistoryMongo() flowHistoryCollection.Aggregate returned error = %v\n", err.Error())
		return -1
	}

	var info []countFlowHistoryType
	err = loadedStructCursor.All(context.TODO(), &info)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countFlowHistoryMongo() loadedStructCursor.All returned error = %v\n", err.Error())
		return -1
	}

	if len(info) == 0 {
		return 0
	}

	return info[0].Count
}

type flowHistoryType struct {
	ID      string `bson:"_id,omitempty"`
	Trigger string `bson:"trigger,omitempty"`
	Count   int    `bson:"count,omitempty"`
	LogData struct {
		JobID   string `bson:"jobID,omitempty"`
		Message string `bson:"message,omitempty"`
	} `bson:"LogData,omitempty"`
}

// return triggers from log based on p
func getFlowHistoryTriggerInfo(p []primitive.D) []flowHistoryType {
	flowHistoryCollection := gfMongoClient.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_LOG_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := flowHistoryCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "getFlowHistoryTriggerInfo() flowHistoryCollection.Aggregate returned error = %v\n", err.Error())
		return nil
	}

	var info []flowHistoryType
	err = loadedStructCursor.All(context.TODO(), &info)
	if err != nil {
		fmt.Fprintf(gfStdErr, "getFlowHistoryTriggerInfo() loadedStructCursor.All returned error = %v\n", err.Error())
		return nil
	}

	return info
}
