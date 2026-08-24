package worker_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/testkit"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"testing"
	"time"
)

type failingPublisher21 struct{}

func (failingPublisher21) Publish(context.Context, string, string, string) error {
	return errors.New("broker down")
}

func TestTask0021PublisherFailureRequeuesEvent(t *testing.T) {
	f := testkit.New(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := f.Store.Exec(context.Background(), "INSERT INTO outbox_messages(id,tenant_id,topic,aggregate_id,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,?)", "event-21", f.TenantID, "topic", "agg", "body", "pending", now, now); err != nil {
		t.Fatal(err)
	}
	r := worker.OutboxRunner{Store: f.Store, Publisher: failingPublisher21{}, Owner: "worker-21", Epoch: 1}
	if err := r.RunOnce(context.Background(), f.TenantID, "event-21"); err == nil {
		t.Fatal("publisher error was hidden")
	}
	var status, owner string
	if err := f.Store.QueryRow(context.Background(), "SELECT status,COALESCE(lease_owner,'') FROM outbox_messages WHERE id='event-21'").Scan(&status, &owner); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || owner != "" {
		t.Fatalf("event not requeued status=%s owner=%s", status, owner)
	}
}
