package flows

import "goflows-processor/flows/goflows"

// ScheduleFlow will create a new trigger (at job) to schedule a flow with cron-expression
func ScheduleFlow(gf *goflows.GoFlow) bool {

	// enable debug logging
	gf.EnableDebug()

	// create the trigger to create the trigger
	// NOTE: UPPERCASE JobInputVar(s) are reserved and removed in "TriggerTriggerSetup"
	myTrigger := goflows.TriggerTriggerSetup(
		goflows.TriggerSetup{
			Name:           gf.GetJobInputVar("NAME"),
			TriggerID:      gf.TriggerID,
			FlowID:         gf.GetJobInputVar("FLOWID"),
			Inputs:         gf.JobFlowInputs,
			CronExpression: gf.GetJobInputVar("CRON-EXPRESSION"),
			RepeatUntil:    gf.GetJobInputVar("REPEAT-UNTIL"),
		})

	gf.Action(
		goflows.FrontEndTriggers, // GoFlows package
		"UpdateTrigger",          // Package function to execute
		myTrigger,
	)

	// disable debug
	gf.DisableDebug()

	return true
}
