package worker_test

import (
	"context"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

func TestTask0010RunningAssignmentRejectsSecondClaim(t *testing.T) {
	f := testkit.New(t)
	tr := f.Train(t)
	c := f.Consignment(t, tr)
	until := time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,lease_owner,lease_epoch,lease_until,next_attempt_at) VALUES(?,?,?,?,?,?,?,?,?)", "assignment-10", f.TenantID, c, "carrier", "running", "worker-b", 2, until, until); err != nil {
		t.Fatal(err)
	}
	w := worker.AssignmentWorker{Store: f.Store}
	if err := w.Claim(context.Background(), f.TenantID, "assignment-10", "worker-a", 3, time.Now().UTC().Add(time.Minute)); err == nil {
		t.Fatal("running assignment was claimed twice")
	}
	var owner string
	_ = f.Store.QueryRow(context.Background(), "SELECT lease_owner FROM route_assignments WHERE id='assignment-10'").Scan(&owner)
	if owner != "worker-b" {
		t.Fatalf("owner=%s", owner)
	}
}
