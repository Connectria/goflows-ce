// dbcallback.go -  goflows callback collection

package goflows

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
  NOTE: the reason that *mongoClient is in the func parameter, is that alert
         actions may need to use when a job is not running or been created
*/

// add or update call back
func UpdateCallBackMongo(mc *mongo.Client, rec CallBackType) error {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))
	opts := options.Update().SetUpsert(true)

	var filter, update bson.M
	filterString := fmt.Sprintf("{\"jobID\": \"%v\"}", rec.JobID)
	_ = bson.UnmarshalExtJSON([]byte(filterString), true, &filter)

	update = bson.M{
		"$set": bson.M{
			"jobID":        rec.JobID,
			"callBackRef":  rec.Reference,
			"callBackID":   rec.ID,
			"callBackData": rec.Data,
			"alertAction":  rec.AlertAction,
		},
	}

	_, err := callBackCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

// delete call back by jobID
func DeleteCallBackByJobID(mc *mongo.Client, jobID string) error {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))

	deleteDoc := bson.M{
		"jobID": jobID,
	}

	_, err := callBackCollection.DeleteOne(context.TODO(), deleteDoc)
	if err != nil {
		return err
	}

	return nil
}

// lookup call back by call back reference
func getCallBackByReference(mc *mongo.Client, ref string) (CallBackType, error) {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))

	lookupDoc := bson.M{
		"callBackRef": ref,
	}

	var found CallBackType

	err := callBackCollection.FindOne(context.TODO(), lookupDoc).Decode(&found)
	if err != nil {
		return CallBackType{}, err
	}

	return found, nil
}

// lookup call back by JobID
func getCallBackByJobID(mc *mongo.Client, ID string) (CallBackType, error) {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))

	lookupDoc := bson.M{
		"jobID": ID,
	}

	var found CallBackType

	err := callBackCollection.FindOne(context.TODO(), lookupDoc).Decode(&found)
	if err != nil {
		return CallBackType{}, err
	}

	return found, nil
}

// lookup call back by ID - this is used by the core processor (that's why it's global)
func GetCallBackByID(mc *mongo.Client, ID string) (CallBackType, error) {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))

	lookupDoc := bson.M{
		"callBackID": ID,
	}

	var found CallBackType

	err := callBackCollection.FindOne(context.TODO(), lookupDoc).Decode(&found)
	if err != nil {
		return CallBackType{}, err
	}

	return found, nil
}

// retrieve all the call backs
func GetCallBacks(mc *mongo.Client, p []bson.D) []CallBackType {
	callBackCollection := mc.Database(os.Getenv("MONGO_GOFLOWS_DATABASE")).Collection(os.Getenv("MONGO_CALLBACK_COLLECTION"))
	timeout, _ := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
	opts := options.Aggregate().SetMaxTime(time.Duration(timeout) * time.Second).SetAllowDiskUse(true)
	loadedStructCursor, err := callBackCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		fmt.Fprintf(gfStdErr, "getCallBack() callBackCollection.Aggregate returned error = %v\n", err.Error())
		return []CallBackType{}
	}

	var cbSet []CallBackType
	err = loadedStructCursor.All(context.TODO(), &cbSet)
	if err != nil {
		fmt.Fprintf(gfStdErr, "getCallBack() loadedStructCursor.All returned error = %v\n", err.Error())
		return []CallBackType{}
	}

	return cbSet
}
