// dbevents.go - database commands relating to eventCollection

package main

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event are Mongo Collection records
type Event struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	SQSMessageID string             `bson:"SQSMessageID"`
	EventData    primitive.D        `bson:"EventData"`
}

// lookupMessageID
func lookupMessageID(messageID string) ([]byte, error) {
	if len(messageID) == 0 {
		logMsg := "messageID is zero in length"
		logger.Error().
			Str("function", "lookupMessageID()").
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	query := fmt.Sprintf("{\"SQSMessageID\": \"%v\"}", messageID)
	var qBson bson.D
	err := bson.UnmarshalExtJSON([]byte(query), true, &qBson)
	if err != nil {
		logMsg := fmt.Sprintf("bson.UnmarshalExtJSON() returns: '%v'", err.Error())
		logger.Error().
			Str("function", "lookupMessageID()").
			Str("messageID", messageID).
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	var bsonDoc bson.D
	err = eventCollection.FindOne(ctx, qBson).Decode(&bsonDoc)
	if err != nil {
		logMsg := fmt.Sprintf("filter = '%v'; eventCollection.FindOne() returns: '%v'", qBson, err.Error())
		logger.Error().
			Str("function", "lookupMessageID()").
			Str("messageID", messageID).
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	return bson.MarshalExtJSON(bsonDoc, true, true)
}

// lookupEventAlertID
func lookupEventAlertID(eventAlertID string) ([]byte, error) {
	if len(eventAlertID) == 0 {
		logMsg := "eventAlertID is zero in length"
		logger.Error().
			Str("function", "lookupEventAlertID()").
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	query := fmt.Sprintf("{\"EventData.alert.alertId\": \"%v\"}", eventAlertID)
	var qBson bson.D
	err := bson.UnmarshalExtJSON([]byte(query), true, &qBson)
	if err != nil {
		logMsg := fmt.Sprintf("bson.UnmarshalExtJSON() returns: '%v'", err.Error())
		logger.Error().
			Str("function", "lookupMessageID()").
			Str("eventAlertID", eventAlertID).
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	var bsonDoc bson.D
	err = eventCollection.FindOne(ctx, qBson).Decode(&bsonDoc)
	if err != nil {
		logMsg := fmt.Sprintf("filter = '%v'; eventCollection.FindOne() returns: '%v'", qBson, err.Error())
		logger.Error().
			Str("function", "lookupMessageID()").
			Str("eventAlertID", eventAlertID).
			Msg(logMsg)
		return nil, errors.New(logMsg)
	}

	return bson.MarshalExtJSON(bsonDoc, true, true)
}
