package goflows

import (
	"regexp"
	"strings"
)

// Regex performs a regular expression on string
func Regex(inputString string, group int, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(inputString)
	if len(match) > 0 {

		// 1 - this is to fix my poor regex skills for iSeries messages
		if strings.HasPrefix(match[group], "/") {
			return strings.TrimPrefix(match[group], "/")
		}

		// 2 - this is to fix my poor regex skills for iSeries messages
		if strings.HasSuffix(match[group], "/") {
			return strings.TrimSuffix(match[group], "/")
		}

		return match[group]
	}

	return ""
}
