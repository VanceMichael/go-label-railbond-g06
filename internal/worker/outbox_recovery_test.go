package worker_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

type failingPublisher struct{ calls int }

func (p *failingPublisher) Publish(context.Context, string, string, string) error {
	p.calls++
	return errors.New("broker down")
}

type succeedAfterRetry struct{ n, okAt int }

func (p *succeedAfterRetry) Publish(context.Context, string, string, string) error {
	p.n++
	if p.n < p.okAt {
		return errors.New("transient")
	}
	return nil
}

func TestOutboxRunnerFailureReleasesLease(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "fail1", f.TenantID, "topic", "aggregate", "body", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	pub := &failingPublisher{}
	r := worker.OutboxRunner{Store: f.Store, Publisher: pub, Owner: "runner", Epoch: 1}
	if err := r.RunOnce(context.Background(), f.TenantID, "fail1"); err == nil {
		t.Fatal("expected publish error")
	}
	var status string
	var leaseOwner sql.NullString
	var lastError sql.NullString
	if err := f.Store.QueryRow(context.Background(), "SELECT status,lease_owner,last_error FROM outbox_messages WHERE id='fail1'").Scan(&status, &leaseOwner, &lastError); err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("expected pending, got %s", status)
	}
	if leaseOwner.Valid && leaseOwner.String != "" {
		t.Fatalf("expected cleared lease, got %q", leaseOwner.String)
	}
	if !lastError.Valid || lastError.String == "" {
		t.Fatal("expected last_error recorded")
	}
}

func TestOutboxRunnerRedeliversAfterFailure(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "fail2", f.TenantID, "topic", "aggregate", "body", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	pub := &succeedAfterRetry{okAt: 2}
	r := worker.OutboxRunner{Store: f.Store, Publisher: pub, Owner: "runner", Epoch: 1}
	_ = r.RunOnce(context.Background(), f.TenantID, "fail2")
	if err := r.RunOnce(context.Background(), f.TenantID, "fail2"); err != nil {
		t.Fatalf("expected redelivery to succeed: %v", err)
	}
	var n int
	if err := f.Store.QueryRow(context.Background(), "SELECT COUNT(*) FROM outbox_messages WHERE id='fail2'").Scan(&n); err != nil || n != 0 {
		t.Fatalf("expected acked message removed, got %d", n)
	}
}
