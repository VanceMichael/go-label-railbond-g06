package worker_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

func TestRuntimeStartStop(t *testing.T) {
	f := testkit.New(t)
	r := &worker.Runtime{Store: f.Store}
	ctx, cancel := context.WithCancel(context.Background())
	r.Start(ctx)
	cancel()
	stop, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	if err := r.Stop(stop); err != nil {
		t.Fatal(err)
	}
}
func TestAssignmentCompletionChecksLease(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,lease_owner,lease_epoch,lease_until,next_attempt_at) VALUES(?,?,?,?,?,?,?,?,?)", "a1", f.TenantID, c, "carrier", "running", "owner", 4, time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	w := worker.AssignmentWorker{Store: f.Store}
	if err := w.Complete(context.Background(), f.TenantID, "a1", "old", 4); err == nil {
		t.Fatal("stale owner completed")
	}
	if err := w.Complete(context.Background(), f.TenantID, "a1", "owner", 4); err != nil {
		t.Fatal(err)
	}
}
func TestRouteCancellationRequeues(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,next_attempt_at) VALUES(?,?,?,?,?,datetime('now'))", "a2", f.TenantID, c, "carrier", "assigned"); err != nil {
		t.Fatal(err)
	}
	f.Store.Hooks.BeforeRouteCall = func(context.Context) error { return context.Canceled }
	err := (&worker.RouteAssignmentWorker{Store: f.Store}).RunOnce(context.Background(), f.TenantID, "a2", "owner", 1)
	if err == nil {
		t.Fatal("carrier cancellation hidden")
	}
	var status string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM route_assignments WHERE id='a2'").Scan(&status); err != nil || status != "retry" {
		t.Fatalf("%v %s", err, status)
	}
}
