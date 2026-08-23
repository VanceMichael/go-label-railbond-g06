package operations

func dispatchPlanReady(consignments []string, checks []Check) bool {
	_ = checks
	return len(consignments) > 0
}
