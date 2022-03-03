// config.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config is the type for passing around configuration
type Config struct {
	AMQprocessor         AMQ    // specific to goflows-processor
	APIPort              string // port used for API
	LogDebug             bool
	MongoGoFlowsDatabase string
	MongoLogCollection   string
	MongoURI             string // mongo server
}

// config
var cfg Config

// database
var logCollection *mongo.Collection
var ctx = context.TODO()

// daemon or cli
var daemonFlag = false

// cron
var scheduler *cron.Cron

// print config
func printConfig() {
	fmt.Printf("Configurable values:\n")
	fmt.Printf("*       AMQprocessor.ClientTag = %v\n", cfg.AMQprocessor.ClientTag)
	fmt.Printf("* AMQprocessor.HistoryExchange = %v\n", cfg.AMQprocessor.HistoryExchange)
	fmt.Printf("*      AMQprocessor.TasksQueue = %v\n", cfg.AMQprocessor.TasksQueue)
	fmt.Printf("*             AMQprocessor.URI = %v\n", cfg.AMQprocessor.URI)
	fmt.Printf("*                      APIPort = %v\n", cfg.APIPort)
	fmt.Printf("*                     LogDebug = %v\n", cfg.LogDebug)
	fmt.Printf("*         MongoGoFlowsDatabase = %v\n", cfg.MongoGoFlowsDatabase)
	fmt.Printf("*           MongoLogCollection = %v\n", cfg.MongoLogCollection)
	fmt.Printf("*                     MongoURI = %v\n", cfg.MongoURI)
	fmt.Printf("\n")
}

// prerequisites before anything
func init() {
	// read config
	err := godotenv.Load()
	if err != nil {

		log.Fatal("ERROR: could not load .env file.")
	}

	// database
	cfg.MongoURI = os.Getenv("MONGO_URI")
	if cfg.MongoURI == "" {
		log.Fatal("Error MONGO_URI not found.")
	}

	cfg.MongoGoFlowsDatabase = os.Getenv("MONGO_GOFLOWS_DATABASE")
	if cfg.MongoGoFlowsDatabase == "" {
		log.Fatal("Error MONGO_GOFLOWS_DATABASE not found.")
	}

	cfg.MongoLogCollection = os.Getenv("MONGO_LOG_COLLECTION")
	if cfg.MongoLogCollection == "" {
		log.Fatal("Error MONGO_LOG_COLLECTION not found.")
	}

	cfg.APIPort = os.Getenv("API_PORT")
	if cfg.APIPort == "" {
		log.Fatal("Error API_PORT not found.")
	}

	// database client
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// RabbitMQ settings
	cfg.AMQprocessor.URI = os.Getenv("AMQ_URI")
	if cfg.AMQprocessor.URI == "" {
		log.Fatal("Error: AMQ_URI not set.")
	}

	cfg.AMQprocessor.ClientTag = os.Getenv("AMQ_CLIENT_TAG")
	if cfg.AMQprocessor.ClientTag == "" {
		log.Fatal("Error: AMQ_CLIENT_TAG not set.")
	}

	cfg.AMQprocessor.HistoryExchange = os.Getenv("AMQ_HISTORY_EXCHANGE")
	if cfg.AMQprocessor.HistoryExchange == "" {
		log.Fatal("Error: AMQ_HISTORY_EXCHANGE not set.")
	}

	cfg.AMQprocessor.TasksQueue = os.Getenv("AMQ_GOFLOWS_SCHEDULER_QUEUE")
	if cfg.AMQprocessor.TasksQueue == "" {
		log.Fatal("Error: AMQ_GOFLOWS_SCHEDULER_QUEUE not set.")
	}

	// logging
	logDebug := strings.ToUpper(os.Getenv("LOG_DEBUG"))
	if logDebug == "" {
		log.Fatal("Error: LOG_DEBUG not set.")
	}

	switch logDebug {
	case "YES":
		cfg.LogDebug = true
	case "NO":
		cfg.LogDebug = false
	default:
		log.Fatal("Error: LOG_DEBUG must be 'Yes' or 'No'")
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.LogDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	mongoWriter := &logToMongoWriter{client}
	logger = zerolog.New(mongoWriter).With().Timestamp().Logger()
}
