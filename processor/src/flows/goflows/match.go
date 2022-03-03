package goflows

import (
	"regexp"
)

// Match checks for a list of possible matches
func (gf *GoFlow) Match(inputVar string, matchList ...string) bool {
	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msgf("Match(inputVar = '%v'; matchlist = '%v')", inputVar, matchList)
	}

	for _, n := range matchList {

		if gf.Debug {
			gf.FlowLogger.Debug().
				Str("jobID", gf.JobID).
				Msgf("regexp.MatchString(n = '%v', inputVar = '%v')", n, inputVar)
		}

		m, err := regexp.MatchString(n, inputVar)
		if m {
			if gf.Debug {
				gf.FlowLogger.Debug().
					Str("jobID", gf.JobID).
					Msg("Match() Returns true.")
			}

			return true
		}

		if err != nil {
			gf.FlowLogger.Warn().
				Str("jobID", gf.JobID).
				Msgf("error with regex ('%v'): '%v'; Match() returns false.", inputVar, err)
			return false
		}
	}

	if gf.Debug {
		gf.FlowLogger.Debug().
			Str("jobID", gf.JobID).
			Msg("Match() Returns false.")
	}
	return false
}
