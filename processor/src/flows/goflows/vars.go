package goflows

import (
	"fmt"
	"unicode"
)

const KEYNOTFOUND = "[key not found]"

// SetFlowVar stores a Key:Value pair
func (gf *GoFlow) SetFlowVar(key, value string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("SetFlowVar(key = '%v'; value ='%v')", key, value)
	}

	gf.FlowVars[key] = value
	return fmt.Sprintf("SetFlowVar('%v')", key)
}

// GetFlowVar returns a Key:Value pair
func (gf *GoFlow) GetFlowVar(key string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GetFlowVar(key = '%v')", key)
	}

	if val, ok := gf.FlowVars[key]; ok {
		if gf.Debug {
			gf.FlowLogger.Debug().
				Str("jobID", gf.JobID).
				Msgf("GetFlowVar(key = '%v') returns '%v'", key, val)
		}
		return val
	}

	returnMsg := KEYNOTFOUND
	gf.FlowLogger.Warn().
		Str("jobID", gf.JobID).
		Msgf("GetFlowVar(key = '%v') returns '%v'", key, returnMsg)
	return returnMsg
}

// InitFlowList is required for FlowListVars (it's a hack)
func (gf *GoFlow) InitFlowList(listName string) {
	gf.FlowListVars[listName] = map[string]interface{}{}
}

// SetFlowList stores a Key:Value pair in a list (dictionary)
func (gf *GoFlow) SetFlowList(listName, key, value string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("SetFlowVarList(listName = '%v'; key = '%v'; value ='%v')", listName, key, value)
	}

	gf.FlowListVars[listName][key] = value
	return fmt.Sprintf("SetFlowListVars('%v','%v')", listName, key)
}

// GetFlowList returns a Key:Value pair from a list (dictionary)
func (gf *GoFlow) GetFlowList(listName, key string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GetFlowVarList(listName = '%v'; key = '%v')", listName, key)
	}

	if val, ok := gf.FlowListVars[listName][key]; ok {
		if gf.Debug {
			gf.FlowLogger.Debug().
				Str("jobID", gf.JobID).
				Msgf("GetFlowVarList(listName = '%v'; key = '%v')", listName, key)
		}
		return val.(string)
	}

	returnMsg := KEYNOTFOUND
	gf.FlowLogger.Warn().
		Str("jobID", gf.JobID).
		Msgf("GetFlowList(listName = '%v'; key = '%v') returns '%v'", listName, key, returnMsg)
	return returnMsg
}

// GetJobInputVar returns a value based on a provided key
func (gf *GoFlow) GetJobInputVar(key string) string {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("GetJobInputVar(key = '%v')", key)
	}

	for _, k := range gf.JobFlowInputs {
		if k.InputName == key {
			if gf.Debug {
				gf.FlowLogger.Debug().
					Str("jobID", gf.JobID).
					Msgf("GetJobInputVar(key = '%v') returns '%v'", key, k.InputValue)
			}
			return k.InputValue
		}
	}

	returnMsg := KEYNOTFOUND
	gf.FlowLogger.Warn().
		Str("jobID", gf.JobID).
		Msgf("GetJobInputVar(key = '%v') returns '%v'", key, returnMsg)
	return returnMsg
}

// refactors a FlowList element to a JSON Array for use in another FlowList element
func (gf *GoFlow) FlowListToJSONArray(key string) []interface{} {
	return []interface{}{gf.FlowListVars[key]}
}

// removeReserved deletes UPPERCASE keys from map
func removeReservedFromJobInputVars(s []FlowInputType) []FlowInputType {

	var tempFlowInputs []FlowInputType
	for _, k := range s {
		if isUpper(k.InputName) {
			continue
		}

		tempFlowInputs = append(tempFlowInputs, k)
	}

	return tempFlowInputs
}

// isUpper checks for uppercase in a string
func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
