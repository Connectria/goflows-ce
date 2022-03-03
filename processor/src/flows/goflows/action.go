package goflows

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Action executes method for GoFlows
func (gf *GoFlow) Action(m map[string]interface{}, name string, params ...interface{}) (results []reflect.Value, err error) {
	var errMsg string
	gf.JobStepID++

	// do not take action of there was an error during GoFlow processing
	if gf.Error {
		err = errors.New("ignoring action due to a previous action error")

		gf.UpdateStatus(
			"Running",                    // runStatus
			"Skipping",                   // stepStatus
			"Action: "+name+" - "+errMsg, // info
			0.0,                          // duration
		)

		gf.FlowLogger.Warn().
			Str("jobID", gf.JobID).
			Int("jobStepID", gf.JobStepID).
			Str("action", name).
			Msgf(errMsg)
		return
	}

	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Int("jobStepID", gf.JobStepID).
			Str("action", name).
			Msgf("Action( m = '%v'; name ='%v'; params = '%v' )", m, name, params)
	}

	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		errMsg = fmt.Sprintf("The number of params supplied (%v) is not adapted (%v)", len(params), f.Type().NumIn())
		err = errors.New(errMsg)
		gf.Error = true

		gf.UpdateStatus(
			"Running",                    // runStatus
			"Failure",                    // stepStatus
			"Action: "+name+" - "+errMsg, // info
			0.0,                          // duration
		)

		gf.FlowLogger.Error().
			Str("jobID", gf.JobID).
			Int("jobStepID", gf.JobStepID).
			Str("action", name).
			Msgf("Action( m = '%v'; name ='%v'; params = '%v' ): %v", m, name, params, errMsg)
		return
	}

	// call the function passing params as arguments
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}

	gf.UpdateStatus(
		"Running",       // runStatus
		"Running",       // stepStatus
		"Action: "+name, // info
		0.0,             // duration
	)

	startTime := time.Now()
	results = f.Call(in)
	duration := time.Since(startTime).Seconds()

	// check to see if the called function returns and 'error' type,
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	for i := 0; i < f.Type().NumOut(); i++ {
		typeOut := f.Type().Out(i)
		switch typeOut.Kind() {
		case reflect.Interface:
			if typeOut.Implements(errorInterface) {
				// TODO: this is a hack and should be rewritten "properly"
				if fmt.Sprintf("%v", results[i]) != "<nil>" {
					err = fmt.Errorf("%v", results[i])
					gf.Error = true
					gf.UpdateStatus(
						"Running",                         // runStatus
						"Failure",                         // stepStatus
						"Action: "+name+" - "+err.Error(), // info
						duration,                          // duration
					)
					gf.FlowLogger.Error().
						Float64("duration", duration).
						Str("jobID", gf.JobID).
						Int("jobStepID", gf.JobStepID).
						Str("action", name).
						Msgf("%v", err.Error())
				}
			}

		default:
			gf.LastActionResults = fmt.Sprintf("%v", results[i])
			gf.UpdateStatus(
				"Running",       // runStatus
				"Success",       // stepStatus
				"Action: "+name, // info
				duration,        // duration
			)
			gf.FlowLogger.Info().
				Float64("duration", duration).
				Str("jobID", gf.JobID).
				Int("jobStepID", gf.JobStepID).
				Str("action", name).
				Msgf("%v", gf.LastActionResults)
		}
	}

	return
}
