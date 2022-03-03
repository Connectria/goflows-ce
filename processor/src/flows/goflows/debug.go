package goflows

// EnableDebug will turn on logger debug on a per GoFlow basis
func (gf *GoFlow) EnableDebug() {
	gf.Debug = true
	gf.FlowLogger.Debug().
		Str("jobID", gf.JobID).
		Msg("EnableDebug()")
}

// DisableDebug will turn off logger debug on a per GoFlow basis
func (gf *GoFlow) DisableDebug() {
	gf.FlowLogger.Debug().
		Str("jobID", gf.JobID).
		Msg("DisableDebug()")
	gf.Debug = false
}
