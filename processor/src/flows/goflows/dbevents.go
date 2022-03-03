// dbevents.go -  opsgenie status collection

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

type countEventType struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Count int64              `bson:"count,omitempty"`
}

// counts of events used in statistic functions
func countEventsMongo(p []primitive.D) int64 {
	eventCollection := gfMongoClient.Database(os.Getenv("MONGO_EVENT_DATABASE")).Collection(os.Getenv("MONGO_EVENT_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := eventCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countEventMongo() eventCollection.Aggregate returned error = %v\n", err.Error())
		return -1
	}

	var info []countEventType
	err = loadedStructCursor.All(context.TODO(), &info)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countEventMongo() eventCollection.All returned error = %v\n", err.Error())
		return -1
	}

	if len(info) == 0 {
		return 0
	}

	return info[0].Count
}
