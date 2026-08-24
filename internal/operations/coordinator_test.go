package operations_test

import (
	"context"
	"testing"

	"github.com/VanceMichael/go-base-railbond-g06/internal/operations"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestBuildDispatchPlanNotReadyWhenCustomsUnreleased(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	id := f.Consignment(t, tr)
	c := operations.Coordinator{Store: f.Store}

	// Booked consignment, customs declaration filed but not yet released.
	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	f.Declaration(t, id, "submitted")

	plan, err := c.BuildDispatchPlan(context.Background(), f.User, tr)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Ready {
		t.Fatalf("plan ready with unreleased customs: %+v", plan.Checks)
	}
	if err := c.CommitDispatch(context.Background(), f.User, plan, "r"); err == nil {
		t.Fatal("commit succeeded on not-ready plan")
	}
}

func TestBuildDispatchPlanReadyWhenCustomsReleased(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	id := f.Consignment(t, tr)
	c := operations.Coordinator{Store: f.Store}

	if _, err := f.Store.Exec(context.Background(), "UPDATE consignments SET status='booked' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	f.Declaration(t, id, "released")

	plan, err := c.BuildDispatchPlan(context.Background(), f.User, tr)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Ready {
		t.Fatalf("plan not ready with released customs: %+v", plan.Checks)
	}
}
