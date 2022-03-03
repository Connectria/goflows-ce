// FLOWS.go - list of goflows to compile into processor

package flows

import (
	"reflect"
	"runtime"
	"strings"

	"goflows-processor/flows/goflows"
)

// GetFuncName returns the basename of the function
func GetFuncName(f interface{}) string {
	strParts := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), ".")
	if len(strParts) > 0 {

		return strParts[1]
	}
	return ""
}

// TaskFlowList are called by triggers
var TaskFlowList = []func(*goflows.GoFlow) bool{
	Dummy,
	ScheduleFlow,
}
