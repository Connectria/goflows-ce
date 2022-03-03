// listFuncs.go	- return the list of compiled goflows

package main

import (
	"fmt"
	"net/http"

	"goflows-processor/flows"

	"github.com/gin-gonic/gin"
)

// list functions as requested via the command line
func listCliFuncs() {
	var tasks int

	fmt.Printf("Task GoFlows functions:\n")
	for _, flow := range flows.TaskFlowList {
		funcName := flows.GetFuncName(flow)
		tasks = tasks + 1
		fmt.Printf("\t%v\n", funcName)
	}

	fmt.Printf("Task GoFlows: %v\n\n", tasks)
}

type subListType struct {
	FuncName string `json:"funcName"`
}

type listType struct {
	TaskFuncs []subListType `json:"taskFuncs,omitempty"`
}

// handleListFuncs
func handleApiListFuncs(c *gin.Context) {
	var taskList []subListType
	for _, flow := range flows.TaskFlowList {
		var taskEntry subListType
		taskEntry.FuncName = flows.GetFuncName(flow)
		taskList = append(taskList, taskEntry)
	}

	var funcs listType
	funcs.TaskFuncs = taskList
	c.JSON(http.StatusOK, gin.H{"funcs": funcs})
}
