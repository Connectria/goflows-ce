// evalRules.go - evaluate an OpsGenie event against event triggers

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"goflows-processor/flows/goflows"

	"github.com/tidwall/gjson"
)

type evalRespStruct struct {
	eventAlertID string
	setVars      map[string]string
	trigger      goflows.TriggerType
}

// evaluate
func evalRules(messageID string) (evalRespStruct, error) {
	logger.Info().
		Str("function", "evalRules()").
		Str("messageID", messageID).
		Msg("messaged received")

	thisEvent, err := lookupMessageID(messageID)
	if err != nil {
		logger.Error().
			Str("function", "evalRules()").
			Str("messageID", messageID).
			Msgf("lookupMessageID returned: '%v'", err.Error())
		return evalRespStruct{}, err
	}

	// unknown event action
	if len(getOpsGenieEventFieldValue(thisEvent, "EventData.action")) == 0 {
		logMsg := "empty OpsGenie event"
		logger.Warn().
			Str("function", "eventRules()").
			Str("messageID", messageID).
			Msg(logMsg)
		return evalRespStruct{}, fmt.Errorf(logMsg)
	}

	// set the alertId used in logging
	thisAlertID := getOpsGenieEventFieldValue(thisEvent, "EventData.alert.alertId")

	// update event trigger cache
	err = triggerEventRules.UpdateEventTriggersCache()
	if err != nil {
		logMsg := fmt.Sprintf("triggerEventRules.UpdateEventTriggersCache() returned: '%v'", err.Error())
		logger.Error().
			Str("function", "evalRules()").
			Str("messageID", messageID).
			Msgf(logMsg)
		return evalRespStruct{}, fmt.Errorf(logMsg)
	}

	// starting event analysis
	startTime := time.Now()
	response := evalRespStruct{
		eventAlertID: thisAlertID,
	}

	logger.Info().
		Str("function", "evalRules()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Msg("starting event trigger analysis")

	// loop through the event triggers
	for _, t := range triggerEventRules.Triggers {

		// reset setVars so one rule doesn't step on another
		response.setVars = make(map[string]string)

		// skip inactive rules
		if !t.Active {
			logger.Warn().
				Str("function", "evalRules()").
				Str("messageID", messageID).
				Str("eventAlertID", thisAlertID).
				Msgf("triggerID('%v') is set to active = false; skipping", t.TriggerID)
			continue
		}

		// rules processing
		eventRulesTotal := len(t.Triggerlogic.EventRules)
		eventRulesMet := 0
		for _, rule := range t.Triggerlogic.EventRules {

			/*
				RULE: comment
					The value begins with two forward slashes (i.e., "//")
					"triggers.triggerLogic.event-rules[]" = "//value"

				EXAMPLE:
				"// Cox email on allmsgws"
				"// Need to extract some vars from a text blob within the event",
				"// set-vars-from-source-regex sets mulitable vars from the same source using seperate regular expressions",

				NOTE: since rule is a byte array, the quote is literally and needs to be escaped
			*/

			if strings.HasPrefix(string(rule), "\"//") {
				eventRulesTotal-- // reduce total since these are just comments
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') - rule found: '//' reducing total rules to %v", t.TriggerID, eventRulesTotal)
				continue
			}

			/*
				RULE: weight	- a way for event rules (triggers) to be prioritized

					1. The higher the number, gets evaluated "sooner"
					2. No number, or 0, will be the default (natural order)
					3. Numbers less than 0 will be evaluated last last, lower the number the "more last it becomes"

				EXAMPLE:
				"weight": 0,	// natural order (default)
				"weight": -1,	// after 0
				"weight": 10,	// before 0 but after 12
				"weight": 12,	// first

				NOTE: skipping these like comments; they are processed when the UpdateEventTriggersCache() is called - sorted at that time.
				This means that by the time they get here, there are already "weighted" in order to be evaluated againste the incoming event.
				For more details, see goflows/triggerEventRules.go on how this is accomplished.

			*/
			ruleExactMatch := getOpsGenieEventFieldBytes(rule, "weight")
			if ruleExactMatch.Index > 0 {
				eventRulesTotal-- // reduce total since these are just comments
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') - rule found: 'weight reducing total rules to %v", t.TriggerID, eventRulesTotal)
				continue
			}

			/*
				RULE: get value match from pattern from alert.
				      "triggers.triggerLogic.event-rules.exact-match" = "value"

				EXAMPLE:
				"exact-match": {"EventData/alert/extraProperties/CustomerID":   "66294"},
				"exact-match": {"EventData/alert/extraProperties/DeviceID":     "11364"},
				"exact-match": {"EventData/alert/extraProperties/service_desc": "Jobs in MSGW"},

			*/
			ruleExactMatch = getOpsGenieEventFieldBytes(rule, "exact-match")
			if ruleExactMatch.Index > 0 {

				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') - rule found: 'exact-match'", t.TriggerID)

				ruleExactMatch.ForEach(
					func(k, v gjson.Result) bool {
						keyDotNotation := strings.Replace(string(k.String()), "/", ".", -1)
						if getOpsGenieEventFieldValue(thisEvent, keyDotNotation) == v.String() {

							eventRulesMet++ // matched
							logger.Info().
								Str("function", "evalRules()").
								Str("messageID", messageID).
								Str("eventAlertID", thisAlertID).
								Msgf("triggerID('%v') event rule matched %v/%v: 'exact-match' key='%v', value='%v'",
									t.TriggerID, eventRulesMet, eventRulesTotal, keyDotNotation, v.String())
						}
						return true
					})
				continue
			}

			/*
				RULE: gf.SetFlowVar from value
				      "triggers.triggerLogic.event-rules.set-var-from-source-value" = "value"

				NOTE: these are not counted as "matching rules" (so each occurance is subtracted from total)

				EXAMPLE:
				"set-var-from-source-value": { "eventAlias": "EventData/alert/extraProperties/alias"}
				"set-var-from-source-value": { "eventTime": "EventData/alert/createdAt"}
			*/

			ruleSetVarFromSourceValue := getOpsGenieEventFieldBytes(rule, "set-var-from-source-value")
			if ruleSetVarFromSourceValue.Index > 0 {

				eventRulesTotal-- // reduce total since these are for reference by other rules
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') event rule found: 'set-var-from-source-value' reducing total rules to %v",
						t.TriggerID, eventRulesTotal)

				ruleSetVarFromSourceValue.ForEach(
					func(k, v gjson.Result) bool {
						valueDotNotation := strings.Replace(v.String(), "/", ".", -1)
						response.setVars[k.String()] = getOpsGenieEventFieldValue(thisEvent, valueDotNotation)
						logger.Info().
							Str("function", "processOpsGenieEvent()").
							Str("messageID", messageID).
							Str("eventAlertID", thisAlertID).
							Msgf("triggerID('%v') event rule 'set-var-from-source-value' key='%v', value='%v', result='%v'",
								t.TriggerID, k.String(), valueDotNotation, response.setVars[k.String()])
						return true
					})
				continue
			}

			/*
				RULE: gf.SetFlowVar using regex expression
				      "triggers.triggerLogic.event-rules.set-vars-from-source-regex"

				NOTE: these are not counted as "matching rules" (so each occurance is subtracted from total)

				SOURCE: "triggers.triggerLogic.event-rules.set-vars-from-source-regex.source"
				VARS:   "triggers.triggerLogic.event-rules.set-vars-from-source-regex.vars[]" = {KEY:VALUE}

				EXAMPLE:
				"set-vars-from-source-regex": {
					"source": "EventData/alert/extraProperties/alias",
				    "vars": [
				        'num':   "(.*?\/){3,3}([^\/]*)",
				        'jobID': "(.*?\/){4,4}([^\/]*)",
				        'msgID': "(.*?\/){5,5}([^\/]*)",
				        'msg':   "(.*?\/){6}(.*[^\/])",
				    ]
			*/
			ruleSetVarsFromSourceRegexSource := getOpsGenieEventFieldValue(rule, "set-vars-from-source-regex.source")
			if len(ruleSetVarsFromSourceRegexSource) > 0 {

				eventRulesTotal-- // reduce total since these are for reference by other rules
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') event rule found: 'set-vars-from-source-regex.source' reducing total rules to %v", t.TriggerID, eventRulesTotal)

				ruleSetVarsFromSourceRegexSourceDotNotation := strings.Replace(ruleSetVarsFromSourceRegexSource, "/", ".", -1)
				logger.Info().
					Str("function", "processOpsGenieEvent()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') event rule: 'set-vars-from-source-regex.source' value='%v'",
						t.TriggerID, ruleSetVarsFromSourceRegexSourceDotNotation)

				regexGroup := 1 // default regex group
				ruleSetVarsFromSourceRegexGroup := getOpsGenieEventFieldValue(rule, "set-vars-from-source-regex.group")
				if len(ruleSetVarsFromSourceRegexGroup) > 0 {
					regexGroup, _ = strconv.Atoi(ruleSetVarsFromSourceRegexGroup)
				}

				ruleSetVarsFromSourceRegexVars := getOpsGenieEventFieldBytes(rule, "set-vars-from-source-regex.vars")
				ruleSetVarsFromSourceRegexVars.ForEach(
					func(_, m gjson.Result) bool {

						m.ForEach(

							func(k, v gjson.Result) bool {
								response.setVars[k.String()] = goflows.Regex(getOpsGenieEventFieldValue(thisEvent, ruleSetVarsFromSourceRegexSourceDotNotation), regexGroup, v.String())
								logger.Info().
									Str("function", "evalRules()").
									Str("messageID", messageID).
									Str("eventAlertID", thisAlertID).
									Msgf("triggerID('%v') event rule: 'set-vars-from-source-regex.vars' key='%v', value='%v', result='%v'",
										t.TriggerID, k.String(), ruleSetVarsFromSourceRegexSourceDotNotation, response.setVars[k.String()])

								return true

							})

						return true
					})

				continue
			}

			/*
				RULE: look for one matching value from list in from the gf.GetFlowVar("name")
				      "triggers.triggerLogic.event-rules.match-all-values-not-in-list

				VAR: "triggers.triggerLogic.event-rules.match-all-values-not-in-list.vars[]"

				NOTE: If the alert key has a matching value, skip/ignore

				EXAMPLE:
				"match-all-values-not-in-list": {
					"jobID": ["DAILYFULL","DAILYINC"],
				}
			*/
			ruleMatchAllValuesNotInList := getOpsGenieEventFieldBytes(rule, "match-all-values-not-in-list")
			if ruleMatchAllValuesNotInList.Index > 0 {
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') - rule found: 'match-all-values-not-in-list'", t.TriggerID)

				valueFound := false
				ruleMatchAllValuesNotInList.ForEach(
					func(k, v gjson.Result) bool {
						v.ForEach(
							func(_, w gjson.Result) bool {

								logger.Debug().
									Str("function", "evalRules()").
									Str("messageID", messageID).
									Str("eventAlertID", thisAlertID).
									Msgf("triggerID('%v') checking event rule for match: - 'match-all-values-not-in-list' key='%v', gf.GetFlowVar returns='%v' w='%v'",
										t.TriggerID, k.String(), response.setVars[k.String()], w.String())

								if response.setVars[k.String()] == w.String() {
									valueFound = true
								}

								return true
							})
						return true
					})

				if !valueFound {
					eventRulesMet++ // we DO NOT WANT to find the value
					logger.Info().
						Str("function", "evalRules()").
						Str("messageID", messageID).
						Str("eventAlertID", thisAlertID).
						Msgf("triggerID('%v') event rule matched %v/%v: - 'match-all-values-not-in-list' list='%v'",
							t.TriggerID, eventRulesMet, eventRulesTotal, ruleMatchAllValuesNotInList)
				}

				continue
			}

			/*
				RULE: look for matching value from list in from the gf.GetFlowVar("name")
				      "triggers.triggerLogic.event-rules.match-all-values-in-list

				VAR: "triggers.triggerLogic.event-rules.match-all-values-not-list.vars[]"

				NOTE: If the alert key has a matching value, true

				EXAMPLE:
				"match-all-values-in-list": {
					"jobID": ["DAILYFULL"],
					"foo": ["bar"],
				}
			*/
			ruleMatchAllValuesInList := getOpsGenieEventFieldBytes(rule, "match-all-values-in-list")
			if ruleMatchAllValuesInList.Index > 0 {
				logger.Info().
					Str("function", "evalRules()").
					Str("messageID", messageID).
					Str("eventAlertID", thisAlertID).
					Msgf("triggerID('%v') - rule found: 'match-all-values-in-list'", t.TriggerID)

				valuesFound := 0
				keyCount := 0
				ruleMatchAllValuesInList.ForEach(
					func(k, v gjson.Result) bool {
						keyCount++
						v.ForEach(
							func(_, w gjson.Result) bool {
								logger.Debug().
									Str("function", "evalRules()").
									Str("messageID", messageID).
									Str("eventAlertID", thisAlertID).
									Msgf("triggerID('%v') checking event rule for match: - 'match-all-values-in-list' key='%v', gf.GetFlowVar returns='%v' w='%v'",
										t.TriggerID, k.String(), response.setVars[k.String()], w.String())

								if response.setVars[k.String()] == w.String() {
									valuesFound++
									logger.Debug().
										Str("function", "evalRules()").
										Str("messageID", messageID).
										Str("eventAlertID", thisAlertID).
										Msgf("triggerid('%v') value found for key: 'match-all-values-in-list' key='%v', gf.getflowvar returns='%v' w='%v'",
											t.TriggerID, k.String(), response.setVars[k.String()], w.String())
								}

								return true
							})
						return true
					})

				if valuesFound >= keyCount {
					eventRulesMet++ // we WANT to find the value
					logger.Info().
						Str("function", "evalRules()").
						Str("messageID", messageID).
						Str("eventAlertID", thisAlertID).
						Msgf("triggerID('%v') event rule matched %v/%v: - 'match-all-values-in-list' list='%v'",
							t.TriggerID, eventRulesMet, eventRulesTotal, ruleMatchAllValuesInList)
				}

				continue
			}
		}

		logger.Info().
			Str("function", "evalRules()").
			Str("messageID", messageID).
			Str("eventAlertID", thisAlertID).
			Msgf("event rules met: %v of %v", eventRulesMet, eventRulesTotal)

		// NOTE: Per SSP-579, we do not want "event rules met: 0 of 0" and flows initiated.
		// if rules match, execute the flows
		if (eventRulesTotal >= 1) && (eventRulesMet == eventRulesTotal) {
			response.trigger = t
			duration := time.Since(startTime).Seconds()
			logger.Info().
				Float64("duration", duration).
				Str("function", "evalRules()").
				Str("messageID", messageID).
				Str("eventAlertID", thisAlertID).
				Str("triggerID", t.TriggerID).
				Msgf("all event rules (criteria) met")
			return response, nil
		}
	}

	// finished
	duration := time.Since(startTime).Seconds()
	logMsg := "finished event rules analysis; no match"
	logger.Info().
		Float64("duration", duration).
		Str("function", "evalRules()").
		Str("messageID", messageID).
		Str("eventAlertID", thisAlertID).
		Msg(logMsg)
	return evalRespStruct{}, fmt.Errorf(logMsg)
}
