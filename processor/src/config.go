// config.go - configuration processing

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"goflows-processor/flows/goflows"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config is the type for passing around configuration
type Config struct {
	AMQprocessor             AMQ           // specific to goflows-processor
	AMQjobStatus             goflows.AMQ   // rabbitMQ for job status
	AlertActionsEnabled      bool          // enable alert actions
	AlertActionsHoursBack    int64         // hours back to display for alert actions
	AlertActionsPort         string        // port for goflows-processor alert actions web server
	AlertActionsProxyURL     string        // reverse proxy for alert actions
	APIPort                  string        // port for goflows-processor API
	AuthLDAPURL              string        // LDAP server used for simple authentication
	AuthTTL                  int64         // time until the token expires
	CallBackPrefixURL        string        // call back Prefix URL
	CallBackInternalURL      string        // call back internal URL
	CallBackPort             string        // call back POST port should be different from APIPort
	FrontEndURL              string        // GoFlows Front End API URL
	FrontEndKey              string        // GoFlows Front End authorization key
	KongAdminAPIURL          string        // Kong Admin Api URL
	MongoAuthCollection      string        // Auth token collection
	MongoCallBackCollection  string        // Call Back collection
	MongoClientTimeout       time.Duration // SetMaxTime option in seconds for MongoDB
	MongoEventDatabase       string        // OpsGenie Events database
	MongoEventCollection     string        // OpsGenie Events collection
	MongoGoFlowsDatabase     string        // GoFlows database
	MongoLogCollection       string        // Logging collection
	MongoReaderLogCollection string        // Logging collection for reader
	MongoStatusCollection    string        // job status collection
	MsgQueueMaxSleep         time.Duration // sleep used by daemon
	MongoURI                 string        // MongoDB
	SMTPServer               string        // mail relay
	TriggerEventRulesTTL     string        // TTL for refreshing TriggerEventRules
}

// default values
const (
	defaultAlertActionsEnabled   = false
	defaultAlertActionsHoursBack = -2
	defaultAlertActionsPort      = "8182"
	defaultMongoClientTimeout    = time.Duration(15) * time.Second  // default is 15 seconds
	defaultSleep                 = time.Duration(300) * time.Second // default is 5 minutes
	defaultWaitTime              = 1
)

// config
var cfg Config

// trigger event rules
var triggerEventRules goflows.TriggerEventRulesCacheType

// database
var eventCollection *mongo.Collection
var logCollection *mongo.Collection
var mongoClient *mongo.Client
var ctx = context.TODO()

// daemon or cli
var daemonFlag = false
var cliRMQ = false

// print config
func printConfig() {
	fmt.Printf("Non-configurable values:\n")
	fmt.Printf("*      defaultAlertActionsEnabled = %v\n", defaultAlertActionsEnabled)
	fmt.Printf("*    defaultAlertActionsHoursBack = %v\n", defaultAlertActionsHoursBack)
	fmt.Printf("*         defaultAlertActionsPort = %v\n", defaultAlertActionsPort)
	fmt.Printf("*                    defaultSleep = %v\n", defaultSleep)
	fmt.Printf("*                 defaultWaitTime = %v\n", defaultWaitTime)
	fmt.Printf("Configurable values:\n")
	fmt.Printf("*             AlertActionsEnabled = %v\n", cfg.AlertActionsEnabled)
	fmt.Printf("*           AlertActionsHoursBack = %v\n", cfg.AlertActionsHoursBack)
	fmt.Printf("*                AlertActionsPort = %v\n", cfg.AlertActionsPort)
	fmt.Printf("*            AlertActionsProxyURL = %v\n", cfg.AlertActionsProxyURL)
	fmt.Printf("*                AMQprocessor.URI = %v\n", cfg.AMQprocessor.URI)
	fmt.Printf("*          AMQprocessor.ClientTag = %v\n", cfg.AMQprocessor.ClientTag)
	fmt.Printf("*    AMQprocessor.HistoryExchange = %v\n", cfg.AMQprocessor.HistoryExchange)
	fmt.Printf("*        AMQprocessor.EventsQueue = %v\n", cfg.AMQprocessor.EventsQueue)
	fmt.Printf("*         AMQprocessor.TasksQueue = %v\n", cfg.AMQprocessor.TasksQueue)
	fmt.Printf("*                AMQjobStatus.URI = %v\n", cfg.AMQjobStatus.URI)
	fmt.Printf("*  AMQjobStatus.JobStatusExchange = %v\n", cfg.AMQjobStatus.JobStatusExchange)
	fmt.Printf("* AMQjobStatus.JobStatusClientTag = %v\n", cfg.AMQjobStatus.JobStatusClientTag)
	fmt.Printf("*                         APIPort = %v\n", cfg.APIPort)
	fmt.Printf("*                         AuthTTL = %v\n", cfg.AuthTTL)
	fmt.Printf("*               CallBackPrefixURL = %v\n", cfg.CallBackPrefixURL)
	fmt.Printf("*             CallBackInternalURL = %v\n", cfg.CallBackInternalURL)
	//fmt.Printf("*                      CallLogURL = %v\n", cfg.CallLogURL)
	//fmt.Printf("*                CallLogPrefixURL = %v\n", cfg.CallLogPrefixURL)
	fmt.Printf("*                     FrontEndURL = %v\n", cfg.FrontEndURL)
	fmt.Printf("*                     FrontEndKey = %v\n", cfg.FrontEndKey)
	fmt.Printf("*                 KongAdminAPIURL = %v\n", cfg.KongAdminAPIURL)
	fmt.Printf("*             MongoAuthCollection = %v\n", cfg.MongoAuthCollection)
	fmt.Printf("*         MongoCallBackCollection = %v\n", cfg.MongoCallBackCollection)
	fmt.Printf("*              MongoClientTimeout = %v\n", cfg.MongoClientTimeout)
	fmt.Printf("*              MongoEventDatabase = %v\n", cfg.MongoEventDatabase)
	fmt.Printf("*            MongoEventCollection = %v\n", cfg.MongoEventCollection)
	fmt.Printf("*            MongoGoFlowsDatabase = %v\n", cfg.MongoGoFlowsDatabase)
	fmt.Printf("*              MongoLogCollection = %v\n", cfg.MongoLogCollection)
	fmt.Printf("*        MongoReaderLogCollection = %v\n", cfg.MongoReaderLogCollection)
	fmt.Printf("*           MongoStatusCollection = %v\n", cfg.MongoStatusCollection)
	fmt.Printf("*                        MongoURI = %v\n", cfg.MongoURI)
	fmt.Printf("*                MsgQueueMaxSleep = %v\n", cfg.MsgQueueMaxSleep)
	fmt.Printf("*                      SMTPServer = %v\n", cfg.SMTPServer)
	fmt.Printf("*            TriggerEventRulesTTL = %v\n", cfg.TriggerEventRulesTTL)
}

// set up
func init() {

	err := godotenv.Load()
	if err != nil {

		log.Fatal("could not load .env file")
	}

	// RabbitMQ settings for GoFlow status publishing
	cfg.AMQjobStatus.JobStatusClientTag = os.Getenv("AMQ_JOBSTATUS_CLIENT_TAG")
	if cfg.AMQjobStatus.JobStatusClientTag == "" {
		log.Fatal("AMQ_JOBSTATUS_CLIENT_TAG not set")
	}

	cfg.AMQjobStatus.JobStatusExchange = os.Getenv("AMQ_JOBSTATUS_EXCHANGE")
	if cfg.AMQjobStatus.JobStatusExchange == "" {
		log.Fatal("AMQ_JOBSTATUS_EXCHANGE not set")
	}

	// RabbitMQ settings
	cfg.AMQprocessor.URI = os.Getenv("AMQ_URI")
	if cfg.AMQprocessor.URI == "" {
		log.Fatal("AMQ_URI not set")
	}
	cfg.AMQjobStatus.URI = cfg.AMQprocessor.URI

	cfg.AMQprocessor.ClientTag = os.Getenv("AMQ_CLIENT_TAG")
	if cfg.AMQprocessor.ClientTag == "" {
		log.Fatal("AMQ_CLIENT_TAG not set")
	}

	cfg.AMQprocessor.HistoryExchange = os.Getenv("AMQ_HISTORY_EXCHANGE")
	if cfg.AMQprocessor.HistoryExchange == "" {
		log.Fatal("AMQ_HISTORY_EXCHANGE not set")
	}

	cfg.AMQprocessor.EventsQueue = os.Getenv("AMQ_OPSGENIE_EVENTS_QUEUE")
	if cfg.AMQprocessor.EventsQueue == "" {
		log.Fatal("AMQ_OPSGENIE_EVENTS_QUEUE not set")
	}

	cfg.AMQprocessor.TasksQueue = os.Getenv("AMQ_GOFLOWS_SCHEDULER_QUEUE")
	if cfg.AMQprocessor.TasksQueue == "" {
		log.Fatal("AMQ_GOFLOWS_SCHEDULER_QUEUE not set")
	}

	// Enable alert actions?
	if os.Getenv("ALERT_ACTIONS_ENABLED") == "YES" {
		cfg.AlertActionsEnabled = true
	} else {
		cfg.AlertActionsEnabled = defaultAlertActionsEnabled
	}

	// Alert Action Hours Back
	if os.Getenv("ALERT_ACTIONS_HOURS_BACK") == "" {
		log.Printf("ALERT_ACTIONS_HOURS_BACK not found. Using default of %v hours back\n", defaultAlertActionsHoursBack)
		cfg.AlertActionsHoursBack = defaultAlertActionsHoursBack
	} else {
		cfg.AlertActionsHoursBack, err = strconv.ParseInt(os.Getenv("ALERT_ACTIONS_HOURS_BACK"), 10, 64)
		if err != nil {
			log.Printf("ALERT_ACTIONS_HOURS_BACK contains an invalid value: %v. Using default of %v hours back\n", os.Getenv("ALERT_ACTIONS_HOURS_BACK"), defaultAlertActionsHoursBack)
			cfg.AlertActionsHoursBack = defaultAlertActionsHoursBack
		}
	}

	// Alert Action port
	cfg.AlertActionsPort = os.Getenv("ALERT_ACTIONS_PORT")
	if cfg.AlertActionsPort == "" {
		log.Printf("ALERT_ACTIONS_PORT not found. Using default port of %v\n", defaultAlertActionsPort)
		cfg.AlertActionsPort = defaultAlertActionsPort
	}

	// Alert Actions reverse proxy URL
	cfg.AlertActionsProxyURL = os.Getenv("ALERT_ACTIONS_PROXY_URL")
	if cfg.AlertActionsProxyURL == "" {
		log.Fatal("ALERT_ACTIONS_PROXY_URL not found")
	}

	// API port
	cfg.APIPort = os.Getenv("API_PORT")
	if cfg.APIPort == "" {
		log.Fatal("API_PORT not found")
	}

	// Call Back Prefix URL - this is the URL to use prefix the Call Back
	cfg.CallBackPrefixURL = os.Getenv("CALLBACK_PREFIX_URL")
	if cfg.CallBackPrefixURL == "" {
		log.Fatal("CALLBACK_PREFIX_URL not found")
	}

	cfg.CallBackInternalURL = os.Getenv("CALLBACK_INTERNAL_URL")
	if cfg.CallBackInternalURL == "" {
		log.Fatal("CALLBACK_INTERNAL_URL not found")
	}

	// Front End API URL
	cfg.FrontEndURL = os.Getenv("FRONTEND_URL")
	if cfg.FrontEndURL == "" {
		log.Fatal("FRONTEND_URL not found")
	}

	// Front End authorization key
	cfg.FrontEndKey = os.Getenv("FRONTEND_KEY")
	if cfg.FrontEndKey == "" {
		log.Fatal("FRONTEND_KEY not found")
	}

	// Kong Admin API
	cfg.KongAdminAPIURL = os.Getenv("KONG_ADMIN_API_URL")
	if cfg.KongAdminAPIURL == "" {
		log.Fatal("KONG_ADMIN_API_URL not set")
	}

	// MongoDB settings
	cfg.MongoURI = os.Getenv("MONGO_URI")
	if cfg.MongoURI == "" {
		log.Fatal("MONGO_URI not set")
	}

	cfg.MongoAuthCollection = os.Getenv("MONGO_AUTH_COLLECTION")
	if cfg.MongoAuthCollection == "" {
		log.Fatal("MONGO_AUTH_COLLECTION not set")
	}

	cfg.MongoCallBackCollection = os.Getenv("MONGO_CALLBACK_COLLECTION")
	if cfg.MongoCallBackCollection == "" {
		log.Fatal("MONGO_CALLBACK_COLLECTION not set")
	}

	cfg.MongoEventDatabase = os.Getenv("MONGO_EVENT_DATABASE")
	if cfg.MongoEventDatabase == "" {
		log.Fatal("MONGO_EVENT_DATABASE not set")
	}

	cfg.MongoEventCollection = os.Getenv("MONGO_EVENT_COLLECTION")
	if cfg.MongoEventCollection == "" {
		log.Fatal("MONGO_EVENT_COLLECTION not set")
	}

	cfg.MongoGoFlowsDatabase = os.Getenv("MONGO_GOFLOWS_DATABASE")
	if cfg.MongoGoFlowsDatabase == "" {
		log.Fatal("MONGO_GOFLOWS_DATABASE not set")
	}

	cfg.MongoLogCollection = os.Getenv("MONGO_LOG_COLLECTION")
	if cfg.MongoLogCollection == "" {
		log.Fatal("MONGO_LOG_COLLECTION not set")
	}

	cfg.MongoReaderLogCollection = os.Getenv("MONGO_READERLOG_COLLECTION")
	if cfg.MongoReaderLogCollection == "" {
		log.Fatal("MONGO_READERLOG_COLLECTION not set")
	}

	cfg.MongoStatusCollection = os.Getenv("MONGO_STATUS_COLLECTION")
	if cfg.MongoStatusCollection == "" {
		log.Fatal("MONGO_STATUS_COLLECTION not set")
	}

	if os.Getenv("MONGO_CLIENT_TIMEOUT") == "" {
		log.Printf("MONGO_CLIENT_TIMEOUT not found. Using default of %v seconds\n", defaultMongoClientTimeout)
		cfg.MongoClientTimeout = defaultMongoClientTimeout
	} else {
		timeout, err := strconv.ParseInt(os.Getenv("MONGO_CLIENT_TIMEOUT"), 10, 64)
		if err != nil {
			log.Printf("MONGO_CLIENT_TIMEOUT contains an invalid value: %v. Using default of %v seconds\n", os.Getenv("MONGO_CLIENT_TIMEOUT"), defaultMongoClientTimeout)
			cfg.MongoClientTimeout = defaultMongoClientTimeout
		} else {
			cfg.MongoClientTimeout = time.Duration(timeout) * time.Second
		}
	}

	// database client
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	eventCollection = mongoClient.Database(cfg.MongoEventDatabase).Collection(cfg.MongoEventCollection)

	// set up logging
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	mongoWriter := &logToMongoWriter{mongoClient}
	logger = zerolog.New(mongoWriter).With().Timestamp().Logger()

	// sleep needed when messages not available
	sleep, err := strconv.ParseInt(os.Getenv("MSGQ_MAX_SLEEP"), 10, 64)
	if err != nil {
		cfg.MsgQueueMaxSleep = defaultSleep
	} else {
		cfg.MsgQueueMaxSleep = time.Duration(sleep) * time.Second
	}

	// mail relay
	cfg.SMTPServer = os.Getenv("SMTP_SERVER")
	if cfg.SMTPServer == "" {
		log.Fatal("SMTP_SERVER not set")
	}

	// TriggerEventRulesTTL
	cfg.TriggerEventRulesTTL = os.Getenv("TRIGGER_EVENT_RULES_TTL")
	if cfg.TriggerEventRulesTTL == "" {
		log.Fatal("TRIGGER_EVENT_RULES_TTL not set")
	}

	// initialize the
	triggerEventRules.LastUpdate = 0
	triggerEventRules.TTL, _ = strconv.ParseInt(cfg.TriggerEventRulesTTL, 10, 64)
	err = triggerEventRules.UpdateEventTriggersCache()
	if err != nil {
		log.Fatalf("%v", err.Error())
	}
}
