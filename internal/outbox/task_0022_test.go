package outbox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/outbox"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
)

func TestTask0022ActiveLeaseRejectsSecondClaim(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC()
	activeUntil := now.Add(time.Minute).Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,lease_owner,lease_epoch,lease_until,attempts,available_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", "event-22", f.TenantID, "topic", "agg", "body", "pending", "worker-a", 1, activeUntil, 1, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}

	service := outbox.Service{Store: f.Store}
	if err := service.Claim(context.Background(), f.User, "event-22", "worker-b", 2, now.Add(2*time.Minute)); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("second claim error = %v, want lease lost", err)
	}
	var status, owner, leaseUntil string
	var epoch, attempts int
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner,lease_epoch,lease_until,attempts FROM outbox_messages WHERE id='event-22'").Scan(&status, &owner, &epoch, &leaseUntil, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || owner != "worker-a" || epoch != 1 || leaseUntil != activeUntil || attempts != 1 {
		t.Fatalf("active lease changed: status=%s owner=%s epoch=%d until=%s attempts=%d", status, owner, epoch, leaseUntil, attempts)
	}

	if _, err := f.Store.Exec(context.Background(), "UPDATE outbox_messages SET lease_until=? WHERE id='event-22'", now.Add(-time.Minute).Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if err := service.Claim(context.Background(), f.User, "event-22", "worker-b", 2, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("expired lease was not claimable: %v", err)
	}
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner,lease_epoch,attempts FROM outbox_messages WHERE id='event-22'").Scan(&status, &owner, &epoch, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "sending" || owner != "worker-b" || epoch != 2 || attempts != 2 {
		t.Fatalf("expired lease claim status=%s owner=%s epoch=%d attempts=%d", status, owner, epoch, attempts)
	}
}
