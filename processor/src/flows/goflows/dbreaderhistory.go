// dbreaderhistory.go -  reader history collection

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

type countReaderHistoryType struct {
	Count int64 `bson:"count,omitempty"`
}

// return count from log based on p for statistic functions
func countReaderHistoryMongo(p []primitive.D) int64 {
	readerHistoryCollection := gfMongoClient.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_READERLOG_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := readerHistoryCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countReaderHistoryMongo() readerHistoryCollection.Aggregate returned error = %v\n", err.Error())
		return -1
	}

	var info []countReaderHistoryType
	err = loadedStructCursor.All(context.TODO(), &info)
	if err != nil {
		fmt.Fprintf(gfStdErr, "countReaderHistoryMongo() loadedStructCursor.All returned error = %v\n", err.Error())
		return -1
	}

	if len(info) == 0 {
		return 0
	}

	return info[0].Count
}
