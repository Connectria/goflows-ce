package goflows

// Criteria are matches to make the flow true
func (gf *GoFlow) Criteria(criteriaList ...interface{}) bool {
	gf.UpdateStatus(
		"Evaluating", // runStatus
		"",           // stepStatus
		gf.FuncName,  // info
		0.0,          // duration
	)

	for _, c := range criteriaList {
		switch c := c.(type) {
		case int:
			if c < 1 {
				return false
			}
		case bool:
			if !c {
				return false
			}
		}
	}

	return true
}
