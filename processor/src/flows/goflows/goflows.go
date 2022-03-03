package goflows

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

// FlowInputType is to an attempt to standardize...?
type FlowInputType struct {
	InputName  string `json:"inputName" bson:"inputName"`   // tag needed to satisfy compiler
	InputValue string `json:"inputValue" bson:"inputValue"` // tag needed to satisfy compiler
}

// GoFlow data structure
type GoFlow struct {
	Debug             bool                              // log debug information
	Error             bool                              // error from previous action
	FlowDescription   string                            // flow description
	FlowLogger        zerolog.Logger                    // where to write the logs
	FlowVars          map[string]string                 // meta: goflow variables
	FlowListVars      map[string]map[string]interface{} // meta: goflow maps(dict/lists)
	FuncName          string                            // goflow function name
	JobControlExit    int64                             // exit: (-1 running, 0 no errors, 1 errors)
	JobCreateTime     int64                             // when the goflow was created
	JobFlowInputs     []FlowInputType                   // inputs to flow
	JobID             string                            // goflow job identifier
	JobName           string                            // goflow job name
	JobSrcFlowID      string                            // flow identififier from api
	JobSrcFlowName    string                            // flow name from api
	JobStepID         int                               // action step
	LastActionResults string                            // last action results
	MessageID         string                            // message identifier
	Tags              []string                          // tags
	TriggerID         string                            // trigger identifier from api
	statusQ           internalAMQ
}

var gfMongoClient *mongo.Client
var gfStdErr, _ = os.OpenFile("goflows-error.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

// New creates a new GoFlow job "run"
func New(messageID string, logger *zerolog.Logger, client *mongo.Client, sq AMQ) *GoFlow {

	// initialize
	gf := &GoFlow{
		FlowVars:     make(map[string]string),
		FlowListVars: make(map[string]map[string]interface{}),
	}

	gf.Debug = false
	gf.Error = false
	gf.FlowDescription = "[flow description not set]"
	gf.FlowLogger = logger.Level(zerolog.DebugLevel)
	gf.JobControlExit = -1
	gf.JobCreateTime = time.Now().Unix()
	gf.JobID = generateUUID()
	gf.JobSrcFlowName = "[flow name not set]"
	gf.JobStepID = 0
	gf.MessageID = messageID
	gf.Tags = make([]string, 0)

	// mongoDB
	gfMongoClient = client

	// RabbitMQ for publishing the job status (actions)
	gf.statusQ.configInfo = sq
	err := gf.initJobStatusQ()
	if err != nil {
		gf.FlowLogger.Error().
			Msgf("initJobStatusQ() returned error: %v", err.Error())
	}

	return gf
}
