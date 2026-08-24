package operations_test

import (
	"context"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/operations"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0028DispatchPlanRejectsFailedPreflight(t *testing.T) {
	f := testkit.New(t)
	trainID := f.Train(t)
	consignmentID := f.Consignment(t, trainID)
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", consignmentID); err != nil {
		t.Fatal(err)
	}
	coordinator := operations.Coordinator{Store: f.Store}
	plan, err := coordinator.BuildDispatchPlan(context.Background(), f.User, trainID)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Consignments) != 1 || plan.Consignments[0] != consignmentID {
		t.Fatalf("dispatch plan consignments=%v", plan.Consignments)
	}
	if plan.Ready {
		t.Fatal("dispatch plan ignored uncleared customs preflight")
	}
	if checkPassed(plan.Checks, "customs_release") {
		t.Fatalf("customs check unexpectedly passed: %#v", plan.Checks)
	}

	f.Declaration(t, consignmentID, "released")
	plan, err = coordinator.BuildDispatchPlan(context.Background(), f.User, trainID)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Ready {
		t.Fatalf("released consignment plan not ready: %#v", plan.Checks)
	}
	for _, check := range plan.Checks {
		if !check.Passed {
			t.Fatalf("ready plan retained failed check: %#v", check)
		}
	}
}

func checkPassed(checks []operations.Check, name string) bool {
	for _, check := range checks {
		if check.Name == name {
			return check.Passed
		}
	}
	return false
}
