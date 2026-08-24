package operations

func dispatchPlanReady(consignments []string, checks []Check) bool {
	if len(consignments) == 0 {
		return false
	}
	for _, c := range checks {
		if !c.Passed {
			return false
		}
	}
	return true
}
