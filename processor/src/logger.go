// logger.go - log writer

package main

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var logger zerolog.Logger

type logToMongoWriter struct {
	client *mongo.Client
}

type logMongoEntryType struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt primitive.DateTime `bson:"createdAt"`
	LogData   primitive.D        `bson:"LogData"`
}

type logRabbitMQType struct {
	Source    string             `json:"source"`
	CreatedAt primitive.DateTime `json:"createdAt"`
	LogData   interface{}        `json:"logdata"`
}

// writer for the logger - overwritten for MongoDB and RabbitMQ
func (ltmw *logToMongoWriter) Write(p []byte) (n int, err error) {

	// create timestamp - needed for MongoDB TTL index
	createdAt := primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond))

	// MongoDB
	logEntry := &logMongoEntryType{
		ID:        primitive.NewObjectID(),
		CreatedAt: createdAt,
	}
	_ = bson.UnmarshalExtJSON(p, true, &logEntry.LogData)
	logCollection = ltmw.client.Database(cfg.MongoGoFlowsDatabase).Collection(cfg.MongoLogCollection)
	_, err = logCollection.InsertOne(ctx, logEntry)
	if err != nil {
		return 0, err
	}

	// RabbitMQ
	if daemonFlag || cliRMQ {
		historyEntry := &logRabbitMQType{
			Source:    "goflows-processor",
			CreatedAt: createdAt,
		}
		_ = json.Unmarshal(p, &historyEntry.LogData)
		postBody, _ := json.Marshal((historyEntry))
		_ = publishLog([]byte(postBody))
	}

	return len(p), nil
}
