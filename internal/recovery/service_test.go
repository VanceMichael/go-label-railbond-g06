package recovery_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/recovery"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestRecoverExpiredRows(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC()
	stamp := now.Add(-time.Hour).Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,lease_until,available_at,created_at) VALUES(?,?,?,?,?,?,?,?,?)", "old", f.TenantID, "x", "a", "p", "sending", stamp, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	r, err := (&recovery.Service{Store: f.Store}).RecoverExpired(context.Background(), now)
	if err != nil || r.OutboxReset != 1 {
		t.Fatalf("%v %#v", err, r)
	}
	var status string
	if err := f.Store.QueryRow(context.Background(), "SELECT status FROM outbox_messages WHERE id='old'").Scan(&status); err != nil || status != "pending" {
		t.Fatalf("outbox not reset: %v %s", err, status)
	}
}

// TestRecoverExpiredAtomicOnFailure asserts the whole recovery batch is atomic:
// when the route-assignment reset fails after the outbox reset ran first inside
// the same transaction, the outbox reset must roll back too. The outbox message
// stays "sending" with its original owner instead of being left half-recovered
// ("pending", ownership lost) in a partial commit.
func TestRecoverExpiredAtomicOnFailure(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC()
	stamp := now.Add(-time.Hour).Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,lease_until,available_at,created_at) VALUES(?,?,?,?,?,?,?,?,?)", "ob", f.TenantID, "x", "a", "p", "sending", stamp, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.Exec(context.Background(), "UPDATE outbox_messages SET lease_owner='w1',lease_epoch=7 WHERE id='ob'"); err != nil {
		t.Fatal(err)
	}
	// Simulate the route-assignment step failing after the outbox reset. WithTx
	// rolls the whole batch back, so neither write commits.
	f.Store.Hooks.BeforeRouteRecovery = func(context.Context) error { return errors.New("route reset failed") }
	t.Cleanup(func() { f.Store.Hooks.BeforeRouteRecovery = nil })

	if _, err := (&recovery.Service{Store: f.Store}).RecoverExpired(context.Background(), now); err == nil {
		t.Fatal("expected rollback error from injected route failure")
	}

	var status string
	var owner sql.NullString
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner FROM outbox_messages WHERE id='ob'").Scan(&status, &owner); err != nil {
		t.Fatal(err)
	}
	if status != "sending" || !owner.Valid || owner.String != "w1" {
		t.Fatalf("outbox partially committed after rollback: status=%s owner=%+v", status, owner)
	}
}
func TestPermanentTargetValidation(t *testing.T) {
	f := testkit.New(t)
	if err := (&recovery.Service{Store: f.Store}).MarkPermanentFailure(context.Background(), f.TenantID, "trains", "x", "bad"); err == nil {
		t.Fatal("arbitrary table accepted")
	}
}
