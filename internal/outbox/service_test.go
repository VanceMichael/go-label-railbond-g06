package outbox_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/domain"
	"github.com/VanceMichael/go-base-railbond-g06/internal/outbox"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"testing"
	"time"
)

func TestClaimAckRequiresOwner(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "o1", f.TenantID, "x", "a", "p", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	s := outbox.Service{Store: f.Store}
	if err := s.Claim(context.Background(), f.User, "o1", "worker-a", 2, time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.Acknowledge(context.Background(), f.User, "o1", "worker-b", 2); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatal(err)
	}
	if err := s.Acknowledge(context.Background(), f.User, "o1", "worker-a", 2); err != nil {
		t.Fatal(err)
	}
}
func TestListOutbox(t *testing.T) {
	f := testkit.New(t)
	s := outbox.Service{Store: f.Store}
	items, err := s.List(context.Background(), f.User)
	if err != nil || items == nil {
		t.Fatalf("%v", err)
	}
}
func TestFailRequiresCurrentOwner(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "o2", f.TenantID, "x", "a", "p", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	s := outbox.Service{Store: f.Store}
	// Worker A claims the message.
	if err := s.Claim(context.Background(), f.User, "o2", "worker-a", 7, time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	// A's lease expires; recovery resets it to pending.
	if _, err := f.Store.Exec(context.Background(), "UPDATE outbox_messages SET status='pending',lease_owner=NULL,lease_until=NULL WHERE id='o2'"); err != nil {
		t.Fatal(err)
	}
	// Worker B reclaims the message with a fresh epoch.
	if err := s.Claim(context.Background(), f.User, "o2", "worker-b", 8, time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	// Stale report from worker A must not clear B's lease.
	if err := s.Fail(context.Background(), f.User, "o2", "worker-a", 7, "stale"); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("expected lease lost, got %v", err)
	}
	var owner string
	var epoch int
	if err := f.Store.QueryRow(context.Background(), "SELECT lease_owner,lease_epoch FROM outbox_messages WHERE id='o2'").Scan(&owner, &epoch); err != nil {
		t.Fatal(err)
	}
	if owner != "worker-b" || epoch != 8 {
		t.Fatalf("lease clobbered: owner=%q epoch=%d", owner, epoch)
	}
	// The current owner may still fail the message.
	if err := s.Fail(context.Background(), f.User, "o2", "worker-b", 8, "real"); err != nil {
		t.Fatalf("%v", err)
	}
}
