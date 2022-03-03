package flows

import "goflows-processor/flows/goflows"

// Dummy
func Dummy(gf *goflows.GoFlow) bool {
	gf.EnableDebug()  // enable debug mode
	gf.DisableDebug() // disable debug mode
	return true
}
