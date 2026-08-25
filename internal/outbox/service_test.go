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
func TestClaimRejectsActiveLease(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "o2", f.TenantID, "x", "a", "p", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	s := outbox.Service{Store: f.Store}
	if err := s.Claim(context.Background(), f.User, "o2", "worker-a", 1, time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if err := s.Claim(context.Background(), f.User, "o2", "worker-b", 2, time.Now().UTC().Add(time.Minute)); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("expected lease lost, got %v", err)
	}
	var owner string
	if err := f.Store.QueryRow(context.Background(), "SELECT lease_owner FROM outbox_messages WHERE tenant_id=? AND id=?", f.TenantID, "o2").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != "worker-a" {
		t.Fatalf("lease owner overwritten to %q", owner)
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
