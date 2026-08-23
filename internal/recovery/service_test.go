package recovery_test

import (
	"context"
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
}
func TestPermanentTargetValidation(t *testing.T) {
	f := testkit.New(t)
	if err := (&recovery.Service{Store: f.Store}).MarkPermanentFailure(context.Background(), f.TenantID, "trains", "x", "bad"); err == nil {
		t.Fatal("arbitrary table accepted")
	}
}
