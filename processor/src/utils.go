// utils.go	- various utility funcs

package main

import (
	"goflows-processor/flows/goflows"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
)

// make this global so it doesn't have to recompile but once
var re = regexp.MustCompile(`(?m)(http|https):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])`)

// find link and make anchor
func link2anchor(str string) string {
	return re.ReplaceAllStringFunc(str, func(s string) string {
		return "<a href='" + s + "'>" + s + "</a>"
	})
}

/*
// TODO: Should this function really just be removed (general cleanup)
// stringify OpsGenie tags
func stringTags(t gjson.Result) string {
	tags := ""
	if t.Index > 0 {
		t.ForEach(
			func(k, v gjson.Result) bool {
				tags = fmt.Sprintf("[%v] %v", v.String(), tags)
				return true
			})
	}

	return tags
}
*/

// array of OpsGenie tags
func arrayTags(t gjson.Result) []string {
	tags := make([]string, 0)
	if t.Index > 0 {
		t.ForEach(
			func(k, v gjson.Result) bool {
				tags = append(tags, v.String())
				return true
			})
	}

	return tags
}

// find JSON patter = return string value
func getOpsGenieEventFieldValue(j []byte, path string) string {
	return gjson.GetBytes(j, path).String()
}

// find JSON pattern = return byte pattern
func getOpsGenieEventFieldBytes(j []byte, path string) gjson.Result {
	return gjson.GetBytes(j, path)
}

// convert the $longnumber(number) into int
func convertCreateAtInt(mStr string) int64 {
	createTimeMillis, _ := strconv.ParseInt(goflows.Regex(mStr, 1, `(?m)([0-9].*\b)`), 10, 64)
	return createTimeMillis / 1000
}

// convert the $longnumber(number) into string
func convertCreateAtStr(mStr string) string {
	createTimeMillis, _ := strconv.ParseInt(goflows.Regex(mStr, 1, `(?m)([0-9].*\b)`), 10, 64)
	createTimeUnix := int(createTimeMillis / 1000)
	return strconv.Itoa(createTimeUnix)
}

/*
// TODO: General Cleanup - while sumBool looks useful is it just extra code at this point?
// this takes an array of booleans and determines if all are true
// useful for finding a match for ALL things
func sumBool(arr []bool) bool {
	sum := 0
	for i := 0; i < len(arr); i++ {
		if arr[i] {
			sum++
		}
	}

	return ((sum / len(arr)) == 1)
}

*/
