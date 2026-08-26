package recovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/recovery"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0027RecoveryFailureRollsBackExpiredLeases(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC()
	expired := now.Add(-time.Minute).Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,lease_owner,lease_epoch,lease_until,available_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)", "outbox-27", f.TenantID, "topic", "aggregate", "body", "sending", "worker-a", 1, expired, expired, expired); err != nil {
		t.Fatal(err)
	}
	consignmentID := f.Consignment(t, f.Train(t))
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO route_assignments(id,tenant_id,consignment_id,carrier,status,lease_owner,lease_epoch,lease_until,next_attempt_at) VALUES(?,?,?,?,?,?,?,?,?)", "assignment-27", f.TenantID, consignmentID, "carrier", "running", "worker-b", 1, expired, expired); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "CREATE TRIGGER fail_assignment_recovery BEFORE UPDATE ON route_assignments BEGIN SELECT RAISE(ABORT, 'assignment recovery unavailable'); END"); err != nil {
		t.Fatal(err)
	}
	if _, err := (&recovery.Service{Store: f.Store}).RecoverExpired(context.Background(), now); err == nil {
		t.Fatal("recovery succeeded while assignment phase failed")
	}
	var outboxStatus, outboxOwner string
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner FROM outbox_messages WHERE id='outbox-27'").Scan(&outboxStatus, &outboxOwner); err != nil {
		t.Fatal(err)
	}
	if outboxStatus != "sending" || outboxOwner != "worker-a" {
		t.Fatalf("failed recovery partially reset outbox status=%s owner=%s", outboxStatus, outboxOwner)
	}
	if _, err := f.Store.Exec(context.Background(), "DROP TRIGGER fail_assignment_recovery"); err != nil {
		t.Fatal(err)
	}
	report, err := (&recovery.Service{Store: f.Store}).RecoverExpired(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	var assignmentStatus string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM route_assignments WHERE id='assignment-27'").Scan(&assignmentStatus); err != nil {
		t.Fatal(err)
	}
	if report.OutboxReset != 1 || report.AssignmentsReset != 1 || assignmentStatus != "retry" {
		t.Fatalf("recovery report=%#v assignment=%s", report, assignmentStatus)
	}
}
